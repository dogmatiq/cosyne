package cosyne

// These methods only exist to prevent the interfaces that declare them from
// being implemented outside this package. They have no behavior to cover.

func (*Latch) sealed()         {}
func (*Future[T]) sealed()     {}
func (*Signal) sealed()        {}
func (*Observable[T]) sealed() {}
