package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/daviddengcn/go-colortext"
	anko_core "github.com/mattn/anko/builtins/core"
	anko_http "github.com/mattn/anko/builtins/http"
	anko_io "github.com/mattn/anko/builtins/io"
	anko_json "github.com/mattn/anko/builtins/json"
	anko_os "github.com/mattn/anko/builtins/os"
	anko_url "github.com/mattn/anko/builtins/url"
	"github.com/mattn/anko/parser"
	"github.com/mattn/anko/vm"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

const version = "0.0.1"

var e = flag.String("e", "", "One line of program")
var verbose = flag.Bool("V", false, "Verbose output")
var v = flag.Bool("v", false, "Display version")

var istty = isatty.IsTerminal(os.Stdout.Fd())

func colortext(color ct.Color, bright bool, f func()) {
	if istty {
		ct.ChangeColor(color, bright, ct.None, false)
		f()
		ct.ResetColor()
	} else {
		f()
	}
}

func main() {
	flag.Parse()
	if *v {
		fmt.Println(version)
		os.Exit(0)
	}

	env := vm.NewEnv()

	anko_core.Import(env)
	anko_http.Import(env)
	anko_url.Import(env)
	anko_json.Import(env)
	anko_os.Import(env)
	anko_io.Import(env)

	if flag.NArg() > 0 || *e != "" {
		var code string
		if *e != "" {
			code = *e
			env.Define("args", reflect.ValueOf(flag.Args()))
		} else {
			body, err := ioutil.ReadFile(flag.Arg(0))
			if err != nil {
				colortext(ct.Red, false, func() {
					fmt.Fprintln(os.Stderr, err)
				})
				os.Exit(1)
			}
			code = string(body)
			env.Define("args", reflect.ValueOf(flag.Args()[1:]))
		}

		scanner := new(parser.Scanner)
		scanner.Init(code)
		stmts, err := parser.Parse(scanner)
		if err != nil {
			colortext(ct.Red, false, func() {
				fmt.Fprintln(os.Stderr, err)
			})
		} else {
			_, err := vm.RunStmts(stmts, env)
			if err != nil {
				colortext(ct.Red, false, func() {
					if e, ok := err.(*vm.Error); ok {
						fmt.Fprintf(os.Stderr, "%s:%d: %s\n", flag.Arg(0), e.Pos().Line, err)
					} else {
						fmt.Fprintln(os.Stderr, err)
					}
				})
				os.Exit(1)
			}
		}
	} else {
		env.Define("args", reflect.ValueOf([]string{}))
		reader := bufio.NewReader(os.Stdin)
		for {
			colortext(ct.Green, true, func() {
				fmt.Print("> ")
			})
			b, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			if len(b) == 0 {
				continue
			}
			s := strings.TrimSpace(string(b))
			scanner := new(parser.Scanner)
			scanner.Init(s)
			stmts, err := parser.Parse(scanner)
			if err != nil {
				colortext(ct.Red, false, func() {
					if e, ok := err.(*vm.Error); ok {
						fmt.Fprintf(os.Stderr, "typein:%d: %s\n", e.Pos().Line, err)
					} else {
						fmt.Fprintln(os.Stderr, err)
					}
				})
			}

			if err == nil {
				v, err := vm.RunStmts(stmts, env)
				if err != nil {
					colortext(ct.Red, false, func() {
						if e, ok := err.(*vm.Error); ok {
							fmt.Fprintf(os.Stderr, "typein:%d: %s\n", e.Pos().Line, err)
						} else {
							fmt.Fprintln(os.Stderr, err)
						}
					})
				} else {
					colortext(ct.Black, true, func() {
						if v == vm.NilValue {
							fmt.Println("nil")
						} else {
							fmt.Println(v.Interface())
						}
					})
				}
			}
			if *verbose {
				env.Dump()
			}
		}
	}
}
