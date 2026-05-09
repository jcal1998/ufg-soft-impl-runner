package br.ufg.inf.assinador;

import br.ufg.inf.assinador.domain.SignRequest;
import br.ufg.inf.assinador.domain.ValidateRequest;
import br.ufg.inf.assinador.domain.SignatureResponse;

public interface SignatureService {
    SignatureResponse sign(SignRequest request);
    SignatureResponse validate(ValidateRequest request);
}
