package util

import "io"

func WriteFull(writer io.Writer, data []byte) error {
	dataLen := len(data)
	written := 0

	for written < dataLen {
		if n, err := writer.Write(data[written:]); err != nil {
			return err
		} else {
			written += n
		}
	}

	return nil
}
