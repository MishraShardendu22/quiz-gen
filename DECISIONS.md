1 using uuid7 over uuid4 cause - UUIDv7 is better than UUIDv4 because its time-ordered makes database inserts significantly more efficient.

2 file folder like design - one topic can have many sub topics and each sub topics cna then have sub sub topics in very details so this is a tree like structure
```type Topic struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type Document struct {
	ID      uuid.UUID `json:"id"`
	TopicID uuid.UUID `json:"topic_id"`
	Name    string    `json:"name"`
	Content string    `json:"content"`
}```

3 if proper md then split the chunks via '##' for a sub topic (this wroks for this i am adding this and only this if there is a scope change then i have coded this in a factory design patteern so we can simply cahnge chunking stretgy like lets say there is a file that we dont know if its proper md or what then we will simply  split it by '\n' after 500 lines +- 100 lines)

4. sync.go design decision, create a hash for every document so that on every restart we can compare the current file with the stored version. If the hash is unchanged, skip re-chunking and re-embedding entirely. If the hash changes, delete the old chunks and regenerate them. If a new file is added, ingest it normally, and if a file is removed from the filesystem, delete its corresponding document and chunks from the database. This keeps the database in sync with the filesystem while avoiding unnecessary work.

