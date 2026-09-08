package system

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const (
	kioskServiceName = "kiosk.service"
	kioskServicePath = "/etc/systemd/system/" + kioskServiceName
)

var kioskUnitTmpl = template.Must(template.New("kiosk").Parse(`[Unit]
Description=DietPi Lightweight Kiosk Display
After=network.target dashboard.service
Wants=dashboard.service

[Service]
User=dietpi
Environment=WPE_COG_PLATFORM=drm
ExecStart=/usr/bin/cog {{.TargetURL}}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`))

// Provisioner inspects the local runtime environment and manages display lifecycle.
type Provisioner struct {
	targetURL string
}

// NewProvisioner instantiates a Provisioner target.
func NewProvisioner(targetURL string) *Provisioner {
	return &Provisioner{
		targetURL: targetURL,
	}
}

func (p *Provisioner) renderUnit() ([]byte, error) {
	var buf bytes.Buffer
	if err := kioskUnitTmpl.Execute(&buf, struct{ TargetURL string }{TargetURL: p.targetURL}); err != nil {
		return nil, fmt.Errorf("failed to execute unit template: %w", err)
	}
	return buf.Bytes(), nil
}

// IsKioskConfigured reports whether the kiosk binary, systemd file, and active status are intact.
func (p *Provisioner) IsKioskConfigured(ctx context.Context) bool {
	if _, err := exec.LookPath("cog"); err != nil {
		return false
	}

	expectedContent, err := p.renderUnit()
	if err != nil {
		return false
	}

	content, err := os.ReadFile(kioskServicePath)
	if err != nil || !bytes.Equal(content, expectedContent) {
		return false
	}

	// Non-zero exit code confirms the unit is not in an active running state
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", "--quiet", kioskServiceName)
	return cmd.Run() == nil
}

// AutoProvision verifies and recovers the display configuration if absent or degraded.
func (p *Provisioner) AutoProvision(ctx context.Context) error {
	if p.IsKioskConfigured(ctx) {
		return nil
	}

	if os.Geteuid() != 0 {
		return errors.New("provisioning requires elevated privileges: run as root or ensure proper sudoer permissions")
	}

	if err := p.installRequiredPackages(ctx); err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	if err := p.writeSystemdService(); err != nil {
		return fmt.Errorf("failed writing systemd unit: %w", err)
	}

	return p.enableAndStartService(ctx)
}

func (p *Provisioner) installRequiredPackages(ctx context.Context) error {
	if _, err := exec.LookPath("cog"); err == nil {
		return nil
	}

	updateCmd := exec.CommandContext(ctx, "apt-get", "update", "-y")
	if out, err := updateCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apt-get update failed: %w: %s", err, string(out))
	}

	installCmd := exec.CommandContext(ctx, "apt-get", "install", "--no-install-recommends", "-y", "cog", "wpewebkit-driver")
	installCmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	if out, err := installCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apt-get install failed: %w: %s", err, string(out))
	}

	return nil
}

func (p *Provisioner) writeSystemdService() error {
	unitContent, err := p.renderUnit()
	if err != nil {
		return err
	}

	existingContent, err := os.ReadFile(kioskServicePath)
	if err == nil && bytes.Equal(existingContent, unitContent) {
		return nil
	}

	return os.WriteFile(kioskServicePath, unitContent, 0644)
}

func (p *Provisioner) enableAndStartService(ctx context.Context) error {
	reloadCmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	if out, err := reloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w: %s", err, string(out))
	}

	enableCmd := exec.CommandContext(ctx, "systemctl", "enable", "--now", kioskServiceName)
	if out, err := enableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable --now: %w: %s", err, string(out))
	}

	return nil
}
