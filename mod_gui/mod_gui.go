package mod_gui

import (
	"strings"
	"strconv"
)

//Pasar el campo directamente desde el array de strings info, y el prefijo que se quiere eliminar.
//Ejemplo: cpu, err := mod_gui.ToFloatFormat(info[0], "CPU: ") 
func toFloatFormat(component string, prefix string) (float64, error) {
    scomponent := strings.TrimPrefix(strings.TrimSpace(component), prefix)
	scomponent = strings.ReplaceAll(scomponent, ",", ".")
	return (strconv.ParseFloat(strings.TrimSpace(scomponent), 64))
}

//Llamar cuando se reciba el mensaje de reporte de consumo del servidor. Retornara los siguientes valores, respectivamente:
//valor de CPU, valor de RAM, valor De Disco, actualización del label de RAM, actualización del label de cantidad de procesos
func ComponentValues(reportMsg string) (cpu float64, ram float64, disk float64, ram_label string, process_label string, err error) {
    info := strings.Split(reportMsg, " | ")
	vcpu, errcpu := toFloatFormat(info[0], "CPU: ") //Valor de la cpu
	vram, _ := toFloatFormat(info[1], "MEM: ") //Valor de la ram
	vram_label := strings.TrimSpace(info[2]) //Actualizacion del label de RAM
	vdisk, _ := toFloatFormat(info[3], "DISK: ") //Valor del disco
    p_label := strings.TrimSpace(info[4]) //Actualizacion del label de cantidad de procesos
	if errcpu != nil {
		return 0.0, 0.0, 0.0, "", "", errcpu
	}
	return vcpu, vram, vdisk, vram_label, p_label, nil
}

