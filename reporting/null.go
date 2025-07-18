package reporting

type NullEmitter struct {
}

func (emitter *NullEmitter) Emit(report *Report) error {
	return nil
}
