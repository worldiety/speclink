// Package amounteditor is a package below the ui directory, not the ui
// directory itself, and is therefore free to be named after what it does.
//
// The presentation layer of a context is regularly more than one package: an
// editor for one widget, a shared table renderer. Demanding uibilling of every
// one of them would force a dozen identically named packages that could only
// be imported through aliases — the very thing the naming rule prevents.
package amounteditor

// Render draws an amount input.
func Render(cents int) string { return "amount" }
