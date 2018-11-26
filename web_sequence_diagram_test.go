package sequence

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWebSequenceDiagram_GeneratesDSL(t *testing.T) {
	wsd := WebSequenceDiagram{}
	wsd.AddRequestRow("A", "B", "request1")
	wsd.AddRequestRow("B", "C", "request2")
	wsd.AddResponseRow("C", "B", "response1")
	wsd.AddResponseRow("B", "A", "response2")

	dsl := wsd.ToString()

	assert.Equal(t, "A->B: (1) request1\nB->C: (2) request2\nC->>B: (3) response1\nB->>A: (4) response2\n", dsl)
}
