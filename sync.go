package markup

import "github.com/murlokswarm/log"

const (
	// FullSync indicates that sync should replace the full node.
	FullSync SyncScope = iota

	// AttrSync indicates that sync should replace only the attributes of the
	// node.
	AttrSync
)

// Sync is a struct which defines how a driver should handle a synchronisation
// of a node on the native side.
type Sync struct {
	Scope      SyncScope
	Index      int
	Node       *Node
	Attributes AttributeMap
}

// SyncScope defines the scope of a sync.
type SyncScope uint8

// Synchronize synchronize a whole component.
// Compares the newer state with the live state of the component.
func Synchronize(c Componer) (syncs []Sync) {
	live := Root(c)

	r, err := render(c)
	if err != nil {
		log.Panic(err)
	}

	new, err := stringToNode(r)
	if err != nil {
		log.Panicf("%T markup returned by Render() has a %v", c, err)
	}

	if new.Type != HTMLNode {
		log.Panicf("%T markup returned by Render() has a syntax error: root node is not a HTMLNode", c)
	}

	syncs, _ = syncNodes(live, new)
	return
}

func syncNodes(live *Node, new *Node) (syncs []Sync, fullSyncParent bool) {
	if live.Type != new.Type {
		replaceNode(live, new)
		fullSyncParent = true
		return
	}

	if live.Type == TextNode {
		if live.Text == new.Text {
			return
		}

		live.Text = new.Text
		fullSyncParent = true
		return
	}

	if live.Tag != new.Tag && live.Type == ComponentNode {
		replaceNode(live, new)
		fullSyncParent = true
		return
	}

	if live.Tag != new.Tag || len(live.Children) != len(new.Children) {
		mergeHTMLNodes(live, new)
		s := Sync{
			Scope: FullSync,
			Node:  live,
		}
		syncs = []Sync{s}
		return
	}

	attrsDiff := live.Attributes.diff(new.Attributes)

	if len(attrsDiff) != 0 && live.Type == ComponentNode {
		live.Attributes = new.Attributes
		decodeAttributeMap(new.Attributes, live.Component)
		syncs = Synchronize(live.Component)
		return
	}

	fullSync := false

	for i := 0; i < len(live.Children); i++ {
		csyncs, fsp := syncNodes(live.Children[i], new.Children[i])
		if !fullSync && fsp {
			fullSync = true
		}

		if fullSync {
			continue
		}

		syncs = append(syncs, csyncs...)
	}

	if fullSync {
		s := Sync{
			Scope: FullSync,
			Node:  live,
		}
		syncs = []Sync{s}
		return
	}

	if len(attrsDiff) != 0 {
		live.Attributes = new.Attributes
		s := Sync{
			Scope:      AttrSync,
			Node:       live,
			Attributes: attrsDiff,
		}
		syncs = append([]Sync{s}, syncs...)
	}
	return
}

func replaceNode(live *Node, new *Node) {
	if live.Type == ComponentNode {
		Dismount(live.Component)
	}

	live.Tag = new.Tag
	live.Type = new.Type
	live.Text = new.Text
	live.Attributes = new.Attributes
	live.Children = new.Children

	for _, c := range live.Children {
		c.Parent = live
	}

	mountNode(live, live.Mount, live.ContextID)
}

func mergeHTMLNodes(live *Node, new *Node) {
	live.Tag = new.Tag
	live.Attributes = new.Attributes

	for _, c := range live.Children {
		dismountNode(c)
	}

	live.Children = new.Children

	for _, c := range live.Children {
		c.Parent = live
		mountNode(c, live.Mount, live.ContextID)
	}
}
