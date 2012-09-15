## A non-working, in-progress Go debugger ##

Eventually this will be a debugger for applications written in the Go
programming language (golang). The debugger itself is written in Go,
and implements interpreters for both Scheme and Tcl. These are used to
write the actual debugging functionality, much as Emacs is largely
implemented in elisp. The inspiration for GoSwat is a debugger for
PC/GEOS called swat, which started life as the debugger for the Sprite
operating system. Naturally it was a Tcl interpreting debugger, since
Sprite and Tcl share a common progenitor.

Presently neither the Scheme nor the Tcl interpreter is complete, nor
is there any debugging functionality. For the time being, you can
check out the code. It would be splendid if GoSwat became something
akin to "living" software...

## From Steve Yegge's [The Pinocchio Problem](http://steve-yegge.blogspot.com/2007/01/pinocchio-problem.html) ##

> Living software has a command shell, since you need a way to talk to
> it like a grown-up. It has an extension language, since you need a
> way to help it grow. It has an advice system, since you need a way
> to train and tailor it. It has a niche, since it needs users in
> order to thrive. It has a plug-in architecture, so you can dress it
> up for your party. And it is self-aware to the maximum extent
> possible given the external performance constraints. These features
> must be seamlessly and elegantly integrated, each subsystem
> implemented with the same care and attention to detail as the system
> as a whole.

## TODO ##

- Make the Scheme and Tcl interpreters work well enough for real work
- Utilize one of the following to facilitate a pseudo GUI in the console:
  - [termbox](https://github.com/nsf/termbox)
  - [gocurse](https://github.com/jabb/gocurse)
  - [goncurses](http://code.google.com/p/goncurses/)
- Evaluate the (now deleted) exp/ogle package to see how it uses debug/elf and debug/dwarf
- Evaluate the (now deleted?) debug/proc package (breakpoints, processes, threads, registers)
- Features
  - Breakpoint tags vs. groups
  - Cool swat features (^u^o)
  - Source view area
  - Variables view area
  - Registers view area
  - Compile Scheme/Tcl to bytecode
