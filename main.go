package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	. "github.com/candid82/joker/core"
	_ "github.com/candid82/joker/std/base64"
	_ "github.com/candid82/joker/std/html"
	_ "github.com/candid82/joker/std/http"
	_ "github.com/candid82/joker/std/json"
	_ "github.com/candid82/joker/std/math"
	_ "github.com/candid82/joker/std/os"
	_ "github.com/candid82/joker/std/string"
	_ "github.com/candid82/joker/std/time"
	_ "github.com/candid82/joker/std/url"
	_ "github.com/candid82/joker/std/yaml"
	"github.com/chzyer/readline"
)

type (
	ReplContext struct {
		first  *Var
		second *Var
		third  *Var
		exc    *Var
	}
)

func NewReplContext(env *Env) *ReplContext {
	first, _ := env.Resolve(MakeSymbol("joker.core/*1"))
	second, _ := env.Resolve(MakeSymbol("joker.core/*2"))
	third, _ := env.Resolve(MakeSymbol("joker.core/*3"))
	exc, _ := env.Resolve(MakeSymbol("joker.core/*e"))
	first.Value = NIL
	second.Value = NIL
	third.Value = NIL
	exc.Value = NIL
	return &ReplContext{
		first:  first,
		second: second,
		third:  third,
		exc:    exc,
	}
}

func (ctx *ReplContext) PushValue(obj Object) {
	ctx.third.Value = ctx.second.Value
	ctx.second.Value = ctx.first.Value
	ctx.first.Value = obj
}

func (ctx *ReplContext) PushException(exc Object) {
	ctx.exc.Value = exc
}

func processFile(filename string, phase Phase) error {
	var reader *Reader
	if filename == "-" || filename == "--" {
		if filename == "--" {
			fmt.Fprintln(os.Stderr, "Warning: '--' indicating standard input (stdin) to Joker is deprecated; please use '-' instead");
		}
		reader = NewReader(bufio.NewReader(os.Stdin), "<stdin>")
		filename = ""
	} else {
		var err error
		reader, err = NewReaderFromFile(filename)
		if err != nil {
			return err
		}
	}
	return ProcessReader(reader, filename, phase)
}

func skipRestOfLine(reader *Reader) {
	for {
		switch reader.Get() {
		case EOF, '\n':
			return
		}
	}
}

func processReplCommand(reader *Reader, phase Phase, parseContext *ParseContext, replContext *ReplContext) (exit bool) {

	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ParseError:
				replContext.PushException(r)
				fmt.Fprintln(os.Stderr, r)
			case *EvalError:
				replContext.PushException(r)
				fmt.Fprintln(os.Stderr, r)
			case Error:
				replContext.PushException(r)
				fmt.Fprintln(os.Stderr, r)
			// case *runtime.TypeAssertionError:
			// 	fmt.Fprintln(os.Stderr, r)
			default:
				panic(r)
			}
		}
	}()

	obj, err := TryRead(reader)
	if err == io.EOF {
		return true
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		skipRestOfLine(reader)
		return
	}

	if phase == READ {
		fmt.Println(obj.ToString(true))
		return false
	}

	expr := Parse(obj, parseContext)
	if phase == PARSE {
		fmt.Println(expr)
		return false
	}

	res := Eval(expr, nil)
	replContext.PushValue(res)
	fmt.Println(res.ToString(true))
	return false
}

func repl(phase Phase) {
	fmt.Printf("Welcome to joker %s. Use EOF (Ctrl-D) or SIGINT (Ctrl-C) to exit.\n", VERSION)
	parseContext := &ParseContext{GlobalEnv: GLOBAL_ENV}
	replContext := NewReplContext(parseContext.GlobalEnv)

	rl, err := readline.New("")
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	defer rl.Close()

	reader := NewReader(NewLineRuneReader(rl), "<repl>")

	for {
		rl.SetPrompt(GLOBAL_ENV.CurrentNamespace().Name.ToString(false) + "=> ")
		if processReplCommand(reader, phase, parseContext, replContext) {
			return
		}
	}
}

func makeDialectKeyword(dialect Dialect) Keyword {
	switch dialect {
	case EDN:
		return MakeKeyword("clj")
	case CLJ:
		return MakeKeyword("clj")
	case CLJS:
		return MakeKeyword("cljs")
	default:
		return MakeKeyword("joker ")
	}
}

func configureLinterMode(dialect Dialect, filename string, workingDir string) {
	ProcessLinterFiles(dialect, filename, workingDir)
	LINTER_MODE = true
	DIALECT = dialect
	lm, _ := GLOBAL_ENV.Resolve(MakeSymbol("joker.core/*linter-mode*"))
	lm.Value = Bool{B: true}
	GLOBAL_ENV.Features = GLOBAL_ENV.Features.Disjoin(MakeKeyword("joker")).Conj(makeDialectKeyword(dialect)).(Set)
	ProcessLinterData(dialect)
}

func detectDialect(filename string) Dialect {
	switch {
	case strings.HasSuffix(filename, ".edn"):
		return EDN
	case strings.HasSuffix(filename, ".cljs"):
		return CLJS
	case strings.HasSuffix(filename, ".joke"):
		return JOKER
	}
	return CLJ
}

func lintFile(filename string, dialect Dialect, workingDir string) {
	phase := PARSE
	if dialect == EDN {
		phase = READ
	}
	ReadConfig(filename, workingDir)
	configureLinterMode(dialect, filename, workingDir)
	if processFile(filename, phase) == nil {
		WarnOnUnusedNamespaces()
		WarnOnUnusedVars()
	}
}

func dialectFromArg(arg string) Dialect {
	switch strings.ToLower(arg) {
	case "clj":
		return CLJ
	case "cljs":
		return CLJS
	case "joker":
		return JOKER
	case "edn":
		return EDN
	}
	return UNKNOWN
}

func usage() {
	fmt.Fprintf(os.Stderr, "Joker - %s\n\n", VERSION)
	fmt.Fprintln(os.Stderr, "Usage: joker [args]                                 starts a repl")
	fmt.Fprintln(os.Stderr, "   or: joker [args] --repl [-- <repl-args>]         starts a repl with args")
	fmt.Fprintln(os.Stderr, "   or: joker [args] --expr <expr> [-- <expr-args>]  input is <expr>")
	fmt.Fprintln(os.Stderr, "   or: joker [args] <filename> [<script-args>]      input from file")
	fmt.Fprintln(os.Stderr, "   or: joker [args] --lint <filename>               lint the code in file")
	fmt.Fprintln(os.Stderr, "\nNotes:")
	fmt.Fprintln(os.Stderr, "  -e is a synonym for --expr")
	fmt.Fprintln(os.Stderr, "  '-' for <filename> means read from standard input (stdin)")
	fmt.Fprintln(os.Stderr, "\nOptions (<args>):")
	fmt.Fprintln(os.Stderr, "  --help, -h")
	fmt.Fprintln(os.Stderr, "    Print this help message and exit.")
	fmt.Fprintln(os.Stderr, "  --version, -v")
	fmt.Fprintln(os.Stderr, "    Print version number and exit.")
	fmt.Fprintln(os.Stderr, "  --read")
	fmt.Fprintln(os.Stderr, "    Read, but do not parse nor evaluate, the input.")
	fmt.Fprintln(os.Stderr, "  --parse")
	fmt.Fprintln(os.Stderr, "    Read and parse, but do not evaluate, the input.")
	fmt.Fprintln(os.Stderr, "  --evaluate")
	fmt.Fprintln(os.Stderr, "    Read, parse, and evaluate the input (default unless --lint in effect).")
	fmt.Fprintln(os.Stderr, "  --working-dir <directory>")
	fmt.Fprintln(os.Stderr, "    Specify working directory for lint configuration (requires --lint).")
	fmt.Fprintln(os.Stderr, "  --dialect <dialect>")
	fmt.Fprintln(os.Stderr, "    Set input dialect (\"clj\", \"cljs\", \"joker\", \"edn\") for linting;")
	fmt.Fprintln(os.Stderr, "    default is inferred from <filename> suffix, if any.")
	fmt.Fprintln(os.Stderr, "  --hashmap-threshold <n>")
	fmt.Fprintln(os.Stderr, "    Set HASHMAP_THRESHOLD accordingly (internal magic of some sort).")
}

var (
	debug bool  // Hidden option
	helpFlag bool
	versionFlag bool
	phase Phase = EVAL
	workingDir string
	lintFlag bool
	dialect Dialect
	expr string
	replFlag bool
	filename string
	remainingArgs []string
)

func parseArgs(args []string) {
	length := len(args)
	stop := false
	missing := false
	noFileFlag := false
	var i int
	for i = 1; i < length; i++ {  // shift
		if (debug) { fmt.Fprintf(os.Stderr, "arg[%d]=%s\n", i, args[i]) }
		switch args[i] {
		case "--", "-":
			stop = true  // "-" is stdin. "--" is stdin for now; later will formally end options processing
		case "--debug":
			debug = true
		case "--help", "-h":
			helpFlag = true
			return  // don't bother parsing anything else
		case "--version", "-v":
			versionFlag = true
		case "--read":
			phase = READ
		case "--parse":
			phase = PARSE
		case "--evaluate":
			phase = EVAL
		case "--working-dir":
			if i < length-1 {
				i += 1  // shift
				workingDir = args[i]
			} else {
				missing = true
			}
		case "--lint":
			lintFlag = true
		case "--lintclj":
			lintFlag = true
			dialect = CLJ
		case "--lintcljs":
			lintFlag = true
			dialect = CLJS
		case "--lintjoker":
			lintFlag = true
			dialect = JOKER
		case "--lintedn":
			lintFlag = true
			dialect = EDN
		case "--dialect":
			if i < length-1 {
				i += 1  // shift
				dialect = dialectFromArg(args[i])
			} else {
				missing = true
			}
		case "--hashmap-threshold":
			if i < length-1 {
				i += 1  // shift
				thresh, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: ", err)
					return
				}
				if thresh < 0 {
					HASHMAP_THRESHOLD = math.MaxInt64
				} else {
					HASHMAP_THRESHOLD = thresh
				}
			} else {
				missing = true
			}
		case "-e", "--expr":
			if i < length-1 {
				i += 1  // shift
				expr = args[i]
				if i < length-1 && args[i+1] == "--" {
					i += 2  // shift 2
					noFileFlag = true
					stop = true
				}
			} else {
				missing = true
			}
		case "--repl":
			replFlag = true
			if i < length-1 && args[i+1] == "--" {
				i += 2  // shift 2
				noFileFlag = true
				stop = true
			}
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "Error: Unrecognized option '%s'\n", args[i])
				os.Exit(2)
			}
			stop = true
		}
		if stop || missing {
			break
		}
	}
	if missing {
		fmt.Fprintf(os.Stderr, "Error: Missing argument for '%s' option\n", args[i])
		os.Exit(3)
	}
	if i < length && !noFileFlag {
		if (debug) { fmt.Fprintf(os.Stderr, "filename=%s\n", args[i]) }
		filename = args[i]
		i += 1  // shift
	}
	if (i < length) {
		if (debug) { fmt.Fprintf(os.Stderr, "remaining=%v\n", args[i:]) }
		remainingArgs = args[i:]
	}
}

func main() {
	GLOBAL_ENV.FindNamespace(MakeSymbol("user")).ReferAll(GLOBAL_ENV.CoreNamespace)

	if os.Args[1] == "--debug" { debug = true }  // peek to see if it's the first arg

	parseArgs(os.Args)

	if debug {
		fmt.Fprintf(os.Stderr, "debug=%v\n", debug)
		fmt.Fprintf(os.Stderr, "helpFlag=%v\n", helpFlag)
		fmt.Fprintf(os.Stderr, "versionFlag=%v\n", versionFlag)
		fmt.Fprintf(os.Stderr, "phase=%v\n", phase)
		fmt.Fprintf(os.Stderr, "lintFlag=%v\n", lintFlag)
		fmt.Fprintf(os.Stderr, "dialect=%v\n", dialect)
		fmt.Fprintf(os.Stderr, "workingDir=%v\n", workingDir)
		fmt.Fprintf(os.Stderr, "HASHMAP_THRESHOLD=%v\n", HASHMAP_THRESHOLD)
		fmt.Fprintf(os.Stderr, "expr=%v\n", expr)
		fmt.Fprintf(os.Stderr, "replFlag=%v\n", replFlag)
		fmt.Fprintf(os.Stderr, "filename=%v\n", filename)
		fmt.Fprintf(os.Stderr, "remainingArgs=%v\n", remainingArgs)
	}

	if (helpFlag) {
		usage()
		return
	}

	if versionFlag {
		println(VERSION)
		return
	}

	if len(remainingArgs) > 0 {
		if lintFlag {
			fmt.Fprintf(os.Stderr, "Error: Cannot provide arguments to code while linting it.\n");
			os.Exit(4)
		}
		if phase != EVAL {
			fmt.Fprintf(os.Stderr, "Error: Cannot provide arguments to code without evaluating it.\n");
			os.Exit(5)
		}
	}

	if expr != "" {
		if lintFlag {
			fmt.Fprintf(os.Stderr, "Error: Cannot combine --expr/-e and --lint.\n");
			os.Exit(6)
		}
		if replFlag {
			fmt.Fprintf(os.Stderr, "Error: Cannot combine --expr/-e and --repl.\n");
			os.Exit(7)
		}
		if workingDir != "" {
			fmt.Fprintf(os.Stderr, "Error: Cannot combine --expr/-e and --working-dir.\n");
			os.Exit(8)
		}
		if filename != "" {
			fmt.Fprintf(os.Stderr, "Error: Cannot combine --expr/-e and a <filename> argument.\n");
			os.Exit(9)
		}
		reader := NewReader(strings.NewReader(expr), "<expr>")
		ProcessReader(reader, "", phase)
		return
	}

	if lintFlag {
		if replFlag {
			fmt.Fprintf(os.Stderr, "Error: Cannot combine --lint and --repl.\n");
			os.Exit(10)
		}
		if dialect == UNKNOWN {
			dialect = detectDialect(filename)
		}
		lintFile(filename, dialect, workingDir)
		if PROBLEM_COUNT > 0 {
			os.Exit(1)
		}
		return
	}

	if workingDir != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify --working-dir option when not linting.\n");
		os.Exit(11)
	}

	if filename != "" {
		processFile(filename, phase)
		return
	}

	repl(phase)
	return
}
