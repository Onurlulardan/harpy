package contract

func (document QueryDocument) Select() RawMember {
	return document.selectMember.clone()
}

func (document QueryDocument) Where() RawMember {
	return document.whereMember.clone()
}

func (document QueryDocument) Sort() RawMember {
	return document.sortMember.clone()
}

func (document QueryDocument) Page() RawMember {
	return document.pageMember.clone()
}

func (document QueryDocument) Search() RawMember {
	return document.searchMember.clone()
}

func (document QueryDocument) Count() RawMember {
	return document.countMember.clone()
}

func (document QueryDocument) Include() RawMember {
	return document.includeMember.clone()
}

func (document QueryDocument) Aggregate() RawMember {
	return document.aggregateMember.clone()
}
