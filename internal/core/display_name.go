package core

// itemName は目録の表示名を返し、未知の ID だけをそのまま返す（DES-WLD-02）。
func (w *World) itemName(id string) string {
	if item, ok := w.catalog.Item(id); ok {
		return item.Name
	}
	return id
}

// gatherName は採取地の表示名を返し、未知の ID だけをそのまま返す（DES-WLD-02）。
func (w *World) gatherName(id string) string {
	if site, ok := w.catalog.Gather(id); ok {
		return site.Name
	}
	return id
}
