package writer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// Update if the file has changed.
func Update(w io.Writer, update bytes.Buffer, file string) error {
	// Check if the file has changed, if not, lets create it for the first time.
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return write(update, file)
	}

	// Load the existing config file so we can compare with the new.
	existing, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "failed to read existing file")
	}

	// Is this the same file?
	if update.String() == string(existing) {
		fmt.Fprintf(w, "File has not changed")
		return nil
	}

	fmt.Fprintf(w, "File has changed. Writing changes.")

	// It is not, lets write update it.
	return write(update, file)
}

// Write the configuration file for HAProxy to consume.
func write(update bytes.Buffer, file string) error {
	// Create a new file which we can apply our template to.
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write to the file.
	_, err = w.Write(update.Bytes())
	if err != nil {
		return err
	}

	return nil
}
