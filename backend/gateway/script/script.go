package script

import "errors"

func Script(scriptName string) error {
	switch scriptName {
	case "example_script":
		return NewExampleScript().Run()
	default:
		return errors.New("script not found")
	}
}
