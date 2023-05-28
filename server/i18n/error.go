package i18n

type Error struct {
  msg *Message
  data map[string]string
}

func NewError(msg *Message, pairs ...string) *Error {
  err := Error{ msg: msg, data: map[string]string{} }
  for i := 0; i < len(pairs) - 1; i += 2 {
    err.data[pairs[i]] = pairs[i + 1]
  }
  return &err
}

func (e *Error) Error() string { return e.msg.Other }

func (e *Error) Localize(localizer *Localizer) string {
  return localizer.Template(e.msg, e.data)
}
