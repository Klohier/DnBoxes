package events

// Aggregate is the base type for domain aggregates that use event sourcing.
// It tracks uncommitted domain events raised during command execution and
// provides version tracking for optimistic concurrency.
type Aggregate struct {
	uncommittedEvents []DomainEvent
	version           int
}

// RecordEvent adds a domain event to the list of uncommitted events.
func (a *Aggregate) RecordEvent(event DomainEvent) {
	a.uncommittedEvents = append(a.uncommittedEvents, event)
}

// IncrementVersion increments the aggregate version after applying an event.
func (a *Aggregate) IncrementVersion() {
	a.version++
}

// UncommittedEvents returns the events raised since the last save.
func (a *Aggregate) UncommittedEvents() []DomainEvent {
	return a.uncommittedEvents
}

// ClearEvents clears uncommitted events after they have been persisted.
func (a *Aggregate) ClearEvents() {
	a.uncommittedEvents = nil
}

// Version returns the current version of the aggregate.
func (a *Aggregate) Version() int {
	return a.version
}
