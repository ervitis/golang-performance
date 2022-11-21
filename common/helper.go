package common

func GetError(chErr <-chan error) error {
	select {
	case err := <-chErr:
		return err
	}
}
