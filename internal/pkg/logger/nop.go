package logger

// Nop возвращает Logger, который ничего не пишет. Полезен в тестах и до настройки логирования.
func Nop() Logger { return nop{} }

type nop struct{}

func (nop) Info(...interface{})                {}
func (nop) Infof(string, ...interface{})       {}
func (nop) Warn(...interface{})                {}
func (nop) Warnf(string, ...interface{})       {}
func (nop) Error(...interface{})               {}
func (nop) Errorf(string, ...interface{})      {}
func (nop) Debug(...interface{})               {}
func (nop) Debugf(string, ...interface{})      {}
func (nop) Fatal(...interface{})               {}
func (nop) Fatalf(string, ...interface{})      {}
func (n nop) WithFields(...interface{}) Logger { return n }
