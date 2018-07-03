package util

import "io"

// WriteFull往指定writer中写数据,知道数据完全写入或者发生错误
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
