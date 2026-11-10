//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/akhilalladi-code/markup-medic/internal/audit"
)

func main() {
	callback := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 {
			return `{"error":"expected elements and metadata"}`
		}
		var elements []audit.Element
		var metadata audit.Metadata
		if err := json.Unmarshal([]byte(args[0].String()), &elements); err != nil {
			return `{"error":"invalid elements payload"}`
		}
		if err := json.Unmarshal([]byte(args[1].String()), &metadata); err != nil {
			return `{"error":"invalid metadata payload"}`
		}
		report, err := json.Marshal(audit.Run(elements, metadata))
		if err != nil {
			return `{"error":"could not encode report"}`
		}
		return string(report)
	})
	js.Global().Set("markupMedicAudit", callback)
	select {}
}
