package domain

type View string

const (
	ViewToday   View = "today"
	ViewInbox   View = "inbox"
	ViewLog     View = "log"
	ViewProject View = "project"
	ViewTrash   View = "trash"
)

const DefaultLogWindowDays = 14
