package web

import "github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"

// SEED dos formularios de briefing (docs §8b).
//
// Desde a Fase 2 (editor), a fonte de verdade dos formularios e o SQLite: a Isa
// edita pelo /admin. Estes structs sao apenas a SEMENTE -- inseridos uma vez, no
// primeiro boot com a tabela `forms` vazia (ver seedDefs -> store.SeedFormsIfEmpty).
// Editar aqui NAO muda um banco ja povoado; para mexer no questionario em
// producao, use o painel.
//
// As perguntas abaixo sao minha leitura do que a Isa pediu: motivacao do ensaio,
// o que a pessoa gosta de comer (ela faz lanche para quem vai ao estudio) e, no
// UGC, por que a marca escolheu ela como creator. O EN e rascunho meu.

// tipoCampo restringe o que o template sabe renderizar. String livre aqui viraria
// um <input type="qualquer-coisa"> silenciosamente errado.
type tipoCampo string

const (
	campoTexto    tipoCampo = "text"
	campoEmail    tipoCampo = "email"
	campoTel      tipoCampo = "tel"
	campoData     tipoCampo = "date"
	campoTextarea tipoCampo = "textarea"
	campoSelect   tipoCampo = "select"
	// campoContato aceita telefone brasileiro OU e-mail -- e o campo por onde a
	// Isa responde, entao errar nele custa o cliente inteiro. Renderiza como
	// text (nao tel: teclado numerico impediria digitar "@"). Ver contatoValido.
	campoContato tipoCampo = "contato"
	// campoArquivo aceita ate maxArquivos imagens. Ver upload.go -- o anexo tem
	// um caminho de dados proprio (disco + tabela submission_files), nao entra
	// no JSON de respostas.
	campoArquivo tipoCampo = "file"
)

// pergunta e um campo do formulario. `Chave` e a identidade: vira o name= do
// input E a chave no JSON salvo. NUNCA renomeie uma chave ja usada em producao --
// as respostas antigas ficariam orfas. Para trocar o texto, mude so o Label.
type pergunta struct {
	Chave       string
	Label       ml
	Ajuda       ml // texto pequeno sob o campo; opcional
	Tipo        tipoCampo
	Obrigatoria bool
	Opcoes      []ml // so para campoSelect
}

// TipoHTML traduz o tipo interno para o `type=` do input. Os tipos proprios
// (campoContato) nao sao valores validos de HTML e precisam virar "text".
func (p pergunta) TipoHTML() string {
	if p.Tipo == campoContato {
		return "text"
	}
	return string(p.Tipo)
}

// formBriefing e um questionario completo.
type formBriefing struct {
	// Slug identifica o form na URL (/briefing/<slug>) e na coluna `form`.
	Slug      string
	Titulo    ml
	Subtitulo ml
	Enviar    ml
	// Sucesso e a mensagem que substitui o form apos o POST (swap do HTMX).
	Sucesso   ml
	Perguntas []pergunta
}

// --- Briefing FOTOGRAFIA -----------------------------------------------------

var briefingFotografia = formBriefing{
	Slug:   "fotografia",
	Titulo: ml{"Vamos combinar seu ensaio", "Let's plan your shoot"},
	Subtitulo: ml{
		"Responde com calma — quanto mais eu souber, melhor eu preparo tudo pra você.",
		"Take your time — the more I know, the better I can prepare everything for you.",
	},
	Enviar: ml{"Enviar briefing", "Send briefing"},
	Sucesso: ml{
		"Recebi seu briefing! 🤍 Vou ler com carinho e te respondo em até 24h úteis.",
		"Got your briefing! 🤍 I'll read it carefully and reply within 24 business hours.",
	},
	Perguntas: []pergunta{
		{
			Chave: "nome", Tipo: campoTexto, Obrigatoria: true,
			Label: ml{"Como você se chama?", "What's your name?"},
		},
		{
			Chave: "contato", Tipo: campoContato, Obrigatoria: true,
			Label: ml{"WhatsApp ou e-mail", "WhatsApp or email"},
			Ajuda: ml{"É por aqui que eu te respondo — confere se está certinho.", "This is how I'll get back to you — please double-check it."},
		},
		{
			Chave: "tipo", Tipo: campoSelect, Obrigatoria: true,
			Label: ml{"Que tipo de trabalho você procura?", "What kind of work are you looking for?"},
			Opcoes: []ml{
				{"Ensaio individual", "Individual shoot"},
				{"Ensaio de casal", "Couple shoot"},
				{"Ensaio temático / conceitual", "Themed / conceptual shoot"},
				{"Evento", "Event"},
				{"Ainda não sei", "Not sure yet"},
			},
		},
		{
			Chave: "data", Tipo: campoData,
			Label: ml{"Tem uma data em mente?", "Do you have a date in mind?"},
			Ajuda: ml{"Se ainda não, tudo bem — deixa em branco.", "If not, that's fine — leave it blank."},
		},
		{
			Chave: "local", Tipo: campoTexto,
			Label: ml{"Cidade ou local", "City or location"},
		},
		{
			// O coracao do pedido da Isa: ela quer entender a PESSOA, nao so agendar.
			Chave: "motivacao", Tipo: campoTextarea, Obrigatoria: true,
			Label: ml{"O que te motivou a querer fazer esse ensaio?", "What made you want to do this shoot?"},
			Ajuda: ml{
				"Pode ser uma fase, uma conquista, um presente, uma vontade antiga — não existe resposta errada.",
				"It could be a phase, an achievement, a gift, a long-held wish — there's no wrong answer.",
			},
		},
		{
			// Campo unico de referencia (antes eram dois, texto + link, em
			// sequencia -- redundante para quem preenche). Link colado aqui
			// vira clicavel no painel; ver autoLink em admin.go.
			Chave: "referencias", Tipo: campoTextarea,
			Label: ml{"Tem referências?", "Any references?"},
			Ajuda: ml{
				"Descreve o clima que você imagina, ou cola links (Pinterest, Drive, perfis).",
				"Describe the mood you picture, or paste links (Pinterest, Drive, profiles).",
			},
		},
		{
			Chave: "referencias_imagens", Tipo: campoArquivo,
			Label: ml{"Ou anexe imagens", "Or attach images"},
			Ajuda: ml{
				"Até 4 imagens, 5 MB cada. Prints da galeria do celular servem perfeitamente.",
				"Up to 4 images, 5 MB each. Screenshots from your phone gallery work perfectly.",
			},
		},
		{
			// A pergunta do lanche. Parece supérflua e e exatamente o oposto: e a
			// assinatura da Isa, o motivo de o briefing nao ser um Google Forms.
			Chave: "lanche", Tipo: campoTexto,
			Label: ml{"O que você gosta de comer?", "What do you like to eat?"},
			Ajuda: ml{
				"Eu preparo um lanchinho pra quem vem ao estúdio ☕ Conta também se tiver alguma restrição ou alergia.",
				"I prepare a little snack for everyone who comes to the studio ☕ Also let me know about any restrictions or allergies.",
			},
		},
		{
			Chave: "observacoes", Tipo: campoTextarea,
			Label: ml{"Mais alguma coisa que eu deva saber?", "Anything else I should know?"},
		},
	},
}

// --- Briefing UGC / CREATOR --------------------------------------------------

var briefingUGC = formBriefing{
	Slug:   "ugc",
	Titulo: ml{"Conta sobre o projeto", "Tell me about the project"},
	Subtitulo: ml{
		"Poucas perguntas pra eu já chegar na nossa conversa com uma proposta na mão.",
		"A few questions so I can come to our conversation with a proposal ready.",
	},
	Enviar: ml{"Enviar briefing", "Send briefing"},
	Sucesso: ml{
		"Briefing recebido! Vou analisar e volto com uma proposta em até 24h úteis.",
		"Briefing received! I'll review it and come back with a proposal within 24 business hours.",
	},
	Perguntas: []pergunta{
		{
			Chave: "marca", Tipo: campoTexto, Obrigatoria: true,
			Label: ml{"Marca / empresa", "Brand / company"},
		},
		{
			Chave: "responsavel", Tipo: campoTexto, Obrigatoria: true,
			Label: ml{"Seu nome e cargo", "Your name and role"},
		},
		{
			Chave: "email", Tipo: campoEmail, Obrigatoria: true,
			Label: ml{"E-mail corporativo", "Work email"},
		},
		{
			Chave: "produto", Tipo: campoTextarea, Obrigatoria: true,
			Label: ml{"Qual produto ou serviço?", "Which product or service?"},
		},
		{
			Chave: "objetivo", Tipo: campoSelect, Obrigatoria: true,
			Label: ml{"Objetivo principal da campanha", "Main goal of the campaign"},
			Opcoes: []ml{
				{"Awareness / alcance", "Awareness / reach"},
				{"Conversão / vendas", "Conversion / sales"},
				{"Lançamento de produto", "Product launch"},
				{"Conteúdo para anúncios (ads)", "Content for ads"},
				{"Prova social / depoimento", "Social proof / testimonial"},
			},
		},
		{
			Chave: "pacote", Tipo: campoSelect,
			Label: ml{"Pacote de interesse", "Package of interest"},
			Ajuda: ml{"Se ainda não decidiu, escolha a última opção.", "If you haven't decided, pick the last option."},
			Opcoes: []ml{
				{"UGC — vídeo avulso", "UGC — single video"},
				{"UGC — combo de vídeos", "UGC — video bundle"},
				{"IGC — publicação no meu perfil", "IGC — post on my profile"},
				{"Quero uma recomendação", "I'd like a recommendation"},
			},
		},
		{
			// A pergunta que a Isa pediu explicitamente. Vale ouro: a resposta diz
			// se a marca conhece o trabalho dela ou esta disparando para 50 creators.
			Chave: "porque", Tipo: campoTextarea, Obrigatoria: true,
			Label: ml{"Por que você me escolheu como creator?", "Why did you choose me as a creator?"},
			Ajuda: ml{
				"Me ajuda muito a entender o que a sua marca enxerga no meu trabalho.",
				"This really helps me understand what your brand sees in my work.",
			},
		},
		{
			Chave: "prazo", Tipo: campoData,
			Label: ml{"Prazo de entrega desejado", "Desired delivery date"},
		},
		{
			Chave: "referencias", Tipo: campoTextarea,
			Label: ml{"Referências ou briefing da campanha", "References or campaign brief"},
			Ajuda: ml{
				"Cola links (Drive, Notion, board) ou descreve o que já tem em mente.",
				"Paste links (Drive, Notion, a board) or describe what you have in mind.",
			},
		},
		{
			Chave: "referencias_imagens", Tipo: campoArquivo,
			Label: ml{"Ou anexe imagens de referência", "Or attach reference images"},
			Ajuda: ml{"Até 4 imagens, 5 MB cada.", "Up to 4 images, 5 MB each."},
		},
		{
			Chave: "observacoes", Tipo: campoTextarea,
			Label: ml{"Outras informações", "Anything else"},
		},
	},
}

// seedDefs converte a semente Go para o shape neutro do store. Chamado uma vez,
// no primeiro boot (ver novoFormsCache). A ordem aqui vira o sort_order inicial.
func seedDefs() []store.FormDef {
	return []store.FormDef{
		formToDef(briefingFotografia, 10),
		formToDef(briefingUGC, 20),
	}
}

func formToDef(f formBriefing, ordem int) store.FormDef {
	d := store.FormDef{
		Slug:        f.Slug,
		TituloPT:    f.Titulo.pt, TituloEN: f.Titulo.en,
		SubtituloPT: f.Subtitulo.pt, SubtituloEN: f.Subtitulo.en,
		EnviarPT:    f.Enviar.pt, EnviarEN: f.Enviar.en,
		SucessoPT:   f.Sucesso.pt, SucessoEN: f.Sucesso.en,
		Visivel:     true,
		Ordem:       ordem,
	}
	for i, p := range f.Perguntas {
		d.Perguntas = append(d.Perguntas, store.QuestionDef{
			Chave:       p.Chave,
			LabelPT:     p.Label.pt, LabelEN: p.Label.en,
			AjudaPT:     p.Ajuda.pt, AjudaEN: p.Ajuda.en,
			Tipo:        string(p.Tipo),
			Obrigatoria: p.Obrigatoria,
			OpcoesJSON:  opcoesToJSON(p.Opcoes),
			Ordem:       (i + 1) * 10,
		})
	}
	return d
}
