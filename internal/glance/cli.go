package glance

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/sensors"
)

type cliIntent uint8

const (
	cliIntentVersionPrint cliIntent = iota
	cliIntentServe
	cliIntentConfigValidate
	cliIntentConfigPrint
	cliIntentDiagnose
	cliIntentSensorsPrint
	cliIntentMountpointInfo
	cliIntentSecretMake
	cliIntentPasswordHash
)

type cliOptions struct {
	intent     cliIntent
	configPath string
	args       []string
}

func parseCliOptions() (*cliOptions, error) {
	var args []string

	args = os.Args[1:]
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v" || args[0] == "version") {
		return &cliOptions{
			intent: cliIntentVersionPrint,
		}, nil
	}

	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println("Usage: glance [options] command")

		fmt.Println("\nOpcje:")
		flags.PrintDefaults()

		fmt.Println("\nPolecenia:")
		fmt.Println("  config:validate       Sprawdzenie poprawności pliku konfiguracyjnego")
		fmt.Println("  config:print          Wyświetlenie sparsowanego pliku konfiguracyjnego z wbudowanymi include'ami")
		fmt.Println("  password:hash <pwd>   Zahashowanie hasła")
		fmt.Println("  secret:make           Wygenerowanie losowego tajnego klucza")
		fmt.Println("  sensors:print         Wyświetlenie wszystkich czujników")
		fmt.Println("  mountpoint:info       Wyświetlenie informacji o danym punkcie montowania")
		fmt.Println("  diagnose              Uruchomienie kontroli diagnostycznych")
	}

	configPath := flags.String("config", "glance.yml", "Set config path")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	var intent cliIntent
	args = flags.Args()
	unknownCommandErr := fmt.Errorf("unknown command: %s", strings.Join(args, " "))

	if len(args) == 0 {
		intent = cliIntentServe
	} else if len(args) == 1 {
		if args[0] == "config:validate" {
			intent = cliIntentConfigValidate
		} else if args[0] == "config:print" {
			intent = cliIntentConfigPrint
		} else if args[0] == "sensors:print" {
			intent = cliIntentSensorsPrint
		} else if args[0] == "diagnose" {
			intent = cliIntentDiagnose
		} else if args[0] == "secret:make" {
			intent = cliIntentSecretMake
		} else {
			return nil, unknownCommandErr
		}
	} else if len(args) == 2 {
		if args[0] == "password:hash" {
			intent = cliIntentPasswordHash
		} else {
			return nil, unknownCommandErr
		}
	} else if len(args) == 2 {
		if args[0] == "mountpoint:info" {
			intent = cliIntentMountpointInfo
		} else {
			return nil, unknownCommandErr
		}
	} else {
		return nil, unknownCommandErr
	}

	return &cliOptions{
		intent:     intent,
		configPath: *configPath,
		args:       args,
	}, nil
}

func cliSensorsPrint() int {
	tempSensors, err := sensors.SensorsTemperatures()
	if err != nil {
		if warns, ok := err.(*sensors.Warnings); ok {
			fmt.Printf("Nie można było pobrać informacji o niektórych czujnikach (%v):\n", err)
			for _, w := range warns.List {
				fmt.Printf(" - %v\n", w)
			}
			fmt.Println()
		} else {
			fmt.Printf("Nie udało się pobrać informacji o czujnikach: %v\n", err)
			return 1
		}
	}

	if len(tempSensors) == 0 {
		fmt.Println("Nie znaleziono czujników")
		return 0
	}

	fmt.Println("Znalezione czujniki:")
	for _, sensor := range tempSensors {
		fmt.Printf(" %s: %.1f°C\n", sensor.SensorKey, sensor.Temperature)
	}

	return 0
}

func cliMountpointInfo(requestedPath string) int {
	usage, err := disk.Usage(requestedPath)
	if err != nil {
		fmt.Printf("Nie udało się pobrać informacji o ścieżce %s: %v\n", requestedPath, err)
		if warns, ok := err.(*disk.Warnings); ok {
			for _, w := range warns.List {
				fmt.Printf(" - %v\n", w)
			}
		}

		return 1
	}

	fmt.Println("Ścieżka:", usage.Path)
	fmt.Println("Typ systemu plików:", ternary(usage.Fstype == "", "unknown", usage.Fstype))
	fmt.Printf("Wykorzystanie: %.1f%%\n", usage.UsedPercent)

	return 0
}
