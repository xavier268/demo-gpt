package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xavier268/demo-openai/config"
)

func main() {

	client := config.NewClient()
	mls, err := client.ListModels(context.Background())
	if err != nil {
		panic(err)
	}
	fname := time.Now().Format("model-2006-01-02.txt")
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "Generated on ", time.Now())
	for _, m := range mls.Models {

		fmt.Fprintf(f, "%20s\t%s\n", m.OwnedBy, m.ID)
	}

}
