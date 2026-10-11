// Throwaway: render profile.
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"waterketen/internal/core"
)

func main() {
	root := core.DataRoot()
	lib, _ := core.OpenLibrary(root)
	atlas, _ := core.NewLibraryAtlas(lib, "")
	names := lib.HasLayers("straten", "water", "riool", "drinkwaterleidingen")
	w, h := 106, 30
	view := core.FitView(atlas.Area.Bounds)
	scene, _ := atlas.Load(view, w, h, names, 0)

	f, _ := os.Create("/tmp/atlas_render2.prof")
	pprof.StartCPUProfile(f)
	start := time.Now()
	for i := 0; i < 10; i++ {
		core.RenderScene(scene, view, w, h, "riool", true, "netwerk", true, true, nil, "", nil, true)
	}
	pprof.StopCPUProfile()
	f.Close()
	fmt.Println("10 renders:", time.Since(start))
}
