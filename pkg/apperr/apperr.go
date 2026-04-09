package apperr

import "errors"

var (
	// ErrNotFound é retornado quando um recurso solicitado não existe.
	ErrNotFound = errors.New("não encontrado")

	// ErrConflict é retornado quando um recurso já existe.
	ErrConflict = errors.New("já existe")

	// ErrUnauthorized é retornado quando as credenciais são inválidas.
	ErrUnauthorized = errors.New("não autorizado")

	// ErrForbidden é retornado quando o usuário não tem permissão para a operação.
	ErrForbidden = errors.New("sem permissão")

	// ErrPendingApproval é retornado quando o cadastro ainda aguarda aprovação.
	ErrPendingApproval = errors.New("cadastro aguardando aprovação")

	// ErrAccountRejected é retornado quando o acesso do usuário foi negado.
	ErrAccountRejected = errors.New("acesso negado pelo administrador")
)
