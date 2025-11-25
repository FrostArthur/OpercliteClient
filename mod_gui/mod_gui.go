package mod_gui

import (
	"strings"
	"strconv"
)

func toFloatFormat(component string, prefix string) (float64, error) {
    scomponent := strings.TrimPrefix(strings.TrimSpace(component), prefix)
	scomponent = strings.ReplaceAll(scomponent, ",", ".")
	return (strconv.ParseFloat(strings.TrimSpace(scomponent), 64))
}

func ComponentValues(reportMsg string) (cpu float64, ram float64, disk float64, ram_label string, process_label string, err error) {
    info := strings.Split(reportMsg, " | ")
	vcpu, errcpu := toFloatFormat(info[0], "CPU: ")
	vram, errram := toFloatFormat(info[1], "MEM: ")
	vram_label := strings.TrimSpace(info[2])
	vdisk, errdisk := toFloatFormat(info[3], "DISK: ")
    p_label := strings.TrimSpace(info[4])
	switch {
	case errcpu != nil:
		return 0.0, 0.0, 0.0, "", "", errcpu
	case errram != nil:
		return 0.0, 0.0, 0.0, "", "", errram
	case errdisk != nil:
		return 0.0, 0.0, 0.0, "", "", errdisk
	}
	return vcpu, vram, vdisk, vram_label, p_label, nil
}

