// Package online is the harness for the tests that need the outside world.
//
// Every test in it sits behind the build constraint `online`, so no ordinary
// build and no ordinary test run compiles one. It is invoked deliberately:
//
//	go test -tags online ./test/online/...
//
// The name says what it needs. It is not integration, not e2e and not slow,
// because none of those tells a reader that running it sends traffic off the
// machine, and that is the only property of it that matters here.
//
// Two rules hold inside it, and they are the reason this package exists rather
// than a build tag scattered through the ordinary suite. It never runs against
// real contributor data: what it resolves is committed here as a fixture, and
// it is never pointed at a corpus an operator supplied or at anything under
// record/. And silence from a harness that did not run must not read like a
// pass, which is why the run reports the number of online tests it ran even
// when that number is zero, and why the gate names this package as not run on
// every one of its own runs.
//
// This file carries no build constraint, so the package exists for every build
// and a reader of the tree finds it. Nothing in it runs anything.
package online
