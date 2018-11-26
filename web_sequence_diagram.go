package sequence

import (
	"bytes"
	"fmt"
)

type WebSequenceDiagram struct {
	data  bytes.Buffer
	count int
}

func (r *WebSequenceDiagram) AddRequestRow(source, target, description string) {
	r.addRow("->", source, target, description)
}

func (r *WebSequenceDiagram) AddResponseRow(source, target, description string) {
	r.addRow("->>", source, target, description)
}

func (r *WebSequenceDiagram) addRow(operation, source, target, description string) {
	r.count += 1
	r.data.WriteString(fmt.Sprintf("%s%s%s: (%d) %s\n",
		source,
		operation,
		target,
		r.count,
		description))
}

func (r *WebSequenceDiagram) ToString() string {
	return r.data.String()
}
