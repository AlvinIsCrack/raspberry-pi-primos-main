package domain

type EnergyPolicy uint8

// Modos de política de energía para los microcontroladores:
// 1 = Normal / Defecto
// 2 = Energy Saver
// 3 = Ultra Energy Saver
const (
	PolicyDefault          EnergyPolicy = 1
	PolicyEnergySaver      EnergyPolicy = 2
	PolicyUltraEnergySaver EnergyPolicy = 3
)

func (p EnergyPolicy) String() string {
	switch p {
	case PolicyDefault:
		return "DEFAULT"
	case PolicyEnergySaver:
		return "ENERGY_SAVER"
	case PolicyUltraEnergySaver:
		return "ULTRA_ENERGY_SAVER"
	default:
		return "UNKNOWN"
	}
}

func (p EnergyPolicy) Byte() byte {
	return byte(p)
}

func (p EnergyPolicy) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
