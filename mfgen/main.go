package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"

	"github.com/urfave/cli"
)

//go:generate go-bindata -o templates.go -prefix ../manifest_templates ../manifest_templates/

const PackageVersion = "4.15.9"

func main() {
	app := cli.NewApp()
	app.Name = "mfgen"
	app.Usage = "Smart Grid Store manifest generator"
	app.Version = PackageVersion
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:   "mksiteconf",
			Usage:  "generate a default site config",
			Action: cli.ActionFunc(genSiteConfig),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out,o",
					Usage: "set the output file",
					Value: "siteconfig.yaml",
				},
			},
		},
		{
			Name:      "generate",
			Usage:     "generate manifests from a site config",
			Action:    cli.ActionFunc(genManifests),
			ArgsUsage: "<siteconfig>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "outdir,o",
					Usage: "set the output directory",
					Value: "manifests",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "proceed even if the manifest directory exists",
				},
			},
		},
		{
			Name:      "custom",
			Usage:     "generate a manifest from a custom template",
			Action:    cli.ActionFunc(genCustom),
			ArgsUsage: "<siteconfig> <template> <outfile>",
			Flags:     []cli.Flag{},
		},
	}
	app.Run(os.Args)
}

func genSiteConfig(c *cli.Context) error {
	f, err := os.OpenFile(c.String("out"), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		fmt.Printf("could not create %q: %v\n", c.String("out"), err)
		os.Exit(1)
	}
	f.WriteString(DefaultSiteConfig)
	f.Close()
	return nil
}

func genCustom(c *cli.Context) error {
	if len(c.Args()) != 3 {
		fmt.Printf("usage: mfgen custom <siteconfig> <template> <outfile>")
		os.Exit(1)
	}
	sc, err := LoadSiteConfig(c.Args()[0])
	if err != nil {
		fmt.Printf("could not load site config: %v\n", err)
		os.Exit(1)
	}
	tsrc, err := ioutil.ReadFile(c.Args()[1])
	if err != nil {
		fmt.Printf("could not load template: %v\n", err)
		os.Exit(1)
	}
	of, err := os.Create(c.Args()[2])
	if err != nil {
		fmt.Printf("could not create output file: %v\n", err)
		os.Exit(1)
	}
	templ, err := template.New("root").Parse(string(tsrc))
	if err != nil {
		fmt.Printf("error: template malformed: %v\n", err)
		os.Exit(1)
	}
	err = templ.Execute(of, sc)
	if err != nil {
		fmt.Printf("error: template error: %v\n", err)
		os.Exit(1)
	}
	of.Close()
	return nil
}

func genManifests(c *cli.Context) error {
	if len(c.Args()) != 1 {
		fmt.Printf("usage: mfgen generate <siteconfig>\n")
		os.Exit(1)
	}
	sc, err := LoadSiteConfig(c.Args()[0])
	if err != nil {
		fmt.Printf("could not load site config: %v\n", err)
		os.Exit(1)
	}
	force := c.Bool("force")
	dir := c.String("outdir")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("could not create directory %q: %v\n", dir, err)
		os.Exit(1)
	}
	if !force {
		//check existing files
		for _, m := range Manifest {
			if !m.Needed(sc) {
				continue
			}
			fn := path.Join(dir, m.Filename)
			if _, err := os.Stat(fn); !os.IsNotExist(err) {
				fmt.Printf("abort! %q exists, pass --force to overwrite\n", fn)
				os.Exit(1)
			}
		}
	}
	for _, m := range Manifest {
		if !m.Needed(sc) {
			continue
		}
		err := m.Generate(dir, sc)
		if err != nil {
			fmt.Printf("abort! could not create %q: %v\n", m.Filename, err)
			os.Exit(1)
		}
	}
	fmt.Printf("manifest files generated successfully\n")
	return nil
}
