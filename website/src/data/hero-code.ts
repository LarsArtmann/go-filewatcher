import { siteConfig } from "./config";

const importPath = siteConfig.github.replace("https://github.com/", "github.com/") + "/v2";

export const heroCode = `package main

import (
    "context"
    "fmt"
    "log"
    "time"

    filewatcher "${importPath}"
)

func main() {
    w, err := filewatcher.New(
        []string{"./src"},
        filewatcher.WithExtensions(".go"),
        filewatcher.WithDebounce(500*time.Millisecond),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer w.Close()

    events, err := w.Watch(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for ev := range events {
        fmt.Printf("%s: %s\\n", ev.Op, ev.Path)
    }
}`;
