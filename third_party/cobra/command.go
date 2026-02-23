package cobra

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Command struct {
	Use   string
	Short string

	Run  func(cmd *Command, args []string)
	RunE func(cmd *Command, args []string) error

	parent      *Command
	children    []*Command
	args        []string
	out         io.Writer
	errOut      io.Writer
	flagSet     *flag.FlagSet
	boolFlags   map[string]*bool
	stringFlags map[string]*string
}

func (c *Command) AddCommand(children ...*Command) {
	for _, child := range children {
		if child == nil {
			continue
		}
		child.parent = c
		c.children = append(c.children, child)
	}
}

func (c *Command) SetArgs(args []string) {
	c.args = append([]string(nil), args...)
}

func (c *Command) SetOut(w io.Writer) {
	c.out = w
}

func (c *Command) SetErr(w io.Writer) {
	c.errOut = w
}

func (c *Command) OutOrStdout() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}

func (c *Command) ErrOrStderr() io.Writer {
	if c.errOut != nil {
		return c.errOut
	}
	return os.Stderr
}

func (c *Command) Println(a ...any) {
	fmt.Fprintln(c.OutOrStdout(), a...)
}

func (c *Command) Printf(format string, a ...any) {
	fmt.Fprintf(c.OutOrStdout(), format, a...)
}

func (c *Command) Flags() *FlagSet {
	if c.flagSet == nil {
		c.flagSet = flag.NewFlagSet(c.name(), flag.ContinueOnError)
		c.flagSet.SetOutput(io.Discard)
		c.boolFlags = make(map[string]*bool)
		c.stringFlags = make(map[string]*string)
	}
	return &FlagSet{
		fs:          c.flagSet,
		boolFlags:   c.boolFlags,
		stringFlags: c.stringFlags,
	}
}

func (c *Command) Execute() error {
	args := c.args
	if args == nil {
		args = os.Args[1:]
	}
	return c.execute(args)
}

func (c *Command) execute(args []string) error {
	if isHelp(args) {
		c.printUsage()
		return nil
	}
	if len(args) > 0 {
		if child := c.findChild(args[0]); child != nil {
			child.out = c.OutOrStdout()
			child.errOut = c.ErrOrStderr()
			return child.execute(args[1:])
		}
	}

	remaining, err := c.parseFlags(args)
	if err != nil {
		return err
	}
	if isHelp(remaining) {
		c.printUsage()
		return nil
	}

	if c.RunE != nil {
		return c.RunE(c, remaining)
	}
	if c.Run != nil {
		c.Run(c, remaining)
		return nil
	}
	if len(c.children) > 0 {
		c.printUsage()
		return nil
	}
	return nil
}

func (c *Command) parseFlags(args []string) ([]string, error) {
	if c.flagSet == nil {
		return args, nil
	}
	if err := c.flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			c.printUsage()
			return nil, nil
		}
		return nil, err
	}
	return c.flagSet.Args(), nil
}

func (c *Command) findChild(name string) *Command {
	for _, child := range c.children {
		if child.name() == name {
			return child
		}
	}
	return nil
}

func (c *Command) name() string {
	use := strings.TrimSpace(c.Use)
	if use == "" {
		return ""
	}
	parts := strings.Fields(use)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func (c *Command) printUsage() {
	name := c.name()
	if name == "" {
		name = "command"
	}
	fmt.Fprintf(c.OutOrStdout(), "Usage: %s\n", c.Use)
	if len(c.children) == 0 {
		return
	}
	fmt.Fprintln(c.OutOrStdout(), "")
	fmt.Fprintln(c.OutOrStdout(), "Available Commands:")
	for _, child := range c.children {
		fmt.Fprintf(c.OutOrStdout(), "  %s\t%s\n", child.name(), child.Short)
	}
}

type FlagSet struct {
	fs          *flag.FlagSet
	boolFlags   map[string]*bool
	stringFlags map[string]*string
}

func (f *FlagSet) StringP(name, shorthand, value, usage string) *string {
	v := f.fs.String(name, value, usage)
	if shorthand != "" {
		f.fs.StringVar(v, shorthand, value, usage)
	}
	f.stringFlags[name] = v
	return v
}

func (f *FlagSet) BoolP(name, shorthand string, value bool, usage string) *bool {
	v := f.fs.Bool(name, value, usage)
	if shorthand != "" {
		f.fs.BoolVar(v, shorthand, value, usage)
	}
	f.boolFlags[name] = v
	return v
}

func (f *FlagSet) GetString(name string) (string, error) {
	v, ok := f.stringFlags[name]
	if !ok {
		return "", fmt.Errorf("flag not found: %s", name)
	}
	return *v, nil
}

func (f *FlagSet) GetBool(name string) (bool, error) {
	v, ok := f.boolFlags[name]
	if !ok {
		return false, fmt.Errorf("flag not found: %s", name)
	}
	return *v, nil
}

func (f *FlagSet) Args() []string {
	return f.fs.Args()
}

func isHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}
