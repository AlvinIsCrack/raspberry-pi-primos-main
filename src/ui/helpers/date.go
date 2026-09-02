package helpers

import (
	"fmt"
	"time"
)

var diasSemana = [...]string{
	time.Sunday:    "Domingo",
	time.Monday:    "Lunes",
	time.Tuesday:   "Martes",
	time.Wednesday: "Miércoles",
	time.Thursday:  "Jueves",
	time.Friday:    "Viernes",
	time.Saturday:  "Sábado",
}

var meses = [...]string{
	time.January:   "Enero",
	time.February:  "Febrero",
	time.March:     "Marzo",
	time.April:     "Abril",
	time.May:       "Mayo",
	time.June:      "Junio",
	time.July:      "Julio",
	time.August:    "Agosto",
	time.September: "Septiembre",
	time.October:   "Octubre",
	time.November:  "Noviembre",
	time.December:  "Diciembre",
}

// FormatSpanishDate formatea una fecha como "Lunes 2 de Septiembre"
func FormatSpanishDate(t time.Time) string {
	diaSemana := diasSemana[t.Weekday()]
	mes := meses[t.Month()]
	return fmt.Sprintf("%s %d de %s", diaSemana, t.Day(), mes)
}
