// Temporary profiling probe for the atlas hot path against the local snapshot.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"waterketen/internal/core"
)

func main() {
	lib, err := core.OpenLibrary(core.DataRoot())
	if err != nil {
		panic(err)
	}
	atlas, err := core.NewLibraryAtlas(lib, "")
	if err != nil {
		panic(err)
	}
	layers := lib.HasLayers("straten", "water", "riool", "drinkwaterleidingen")
	view := core.FitView(atlas.Area.Bounds)
	scene, err := atlas.Load(view, 106, 30, layers, 0)
	if err != nil {
		panic(err)
	}
	runtime.MemProfileRate = 1
	for range 10 {
		core.RenderScene(scene, view, 106, 30, "riool", true, "netwerk", true, true, nil, "", nil, true)
	}
	f, err := os.Create("/home/hermes/.hermes/cache/scratch/wk-allocs.prof")
	if err != nil {
		panic(err)
	}
	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		panic(err)
	}
	_ = f.Close()
	fmt.Println("allocation profile saved")
}
