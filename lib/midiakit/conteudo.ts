export type Par = readonly [pt: string, en: string];

export type Pacote = { nome: Par; resumo: Par; itens: Par[]; preco: string; destaque?: boolean };
export type BlocoPacotes = { titulo: Par; subtitulo?: Par; tag: string; pacotes: Pacote[] };
export type Passo = { num: string; titulo: Par; resumo: Par; prazo: Par; detalhes: Par[] };
export type Barra = { rotulo: Par; pct: number };
export type Fatia = { rotulo: Par; pct: number; cor: string };
export type Rede = {
  rede: string;
  seguidores: string;
  views: string;
  locais: Barra[];
  genero: Fatia[];
  idades: Fatia[];
  foto: string;
  lado: string;
};
export type Video = { titulo: string; src: string; poster: string };
export type Categoria = { slug: string; nome: string; videos: Video[] };

export const TAGS: Par[] = [
  [
    "Storytelling",
    "Storytelling"
  ],
  [
    "Direção de Arte",
    "Art Direction"
  ],
  [
    "Cinema",
    "Film"
  ],
  [
    "Edição Caprichada",
    "Meticulous Editing"
  ]
];

export const TEMPO_MEDIO: Par = [
  "Tempo médio: 7–12 dias (dependendo da demanda)",
  "Average time: 7–12 days (depending on demand)"
];

export const INVESTIMENTOS: BlocoPacotes[] = [
  {
    "titulo": [
      "Investimento UGC",
      "UGC Investment"
    ],
    "subtitulo": [
      "Escolha a jornada ideal para sua marca.",
      "Choose the ideal journey for your brand."
    ],
    "tag": "UGC",
    "pacotes": [
      {
        "nome": [
          "Pacote 1",
          "Package 1"
        ],
        "resumo": [
          "1 vídeo vertical",
          "1 vertical video"
        ],
        "preco": "R$ 300",
        "itens": [
          [
            "Roteiro simples",
            "Simple script"
          ],
          [
            "1 foto",
            "1 photo"
          ],
          [
            "Edição padrão",
            "Standard editing"
          ],
          [
            "Legendas",
            "Captions"
          ],
          [
            "1 revisão simples",
            "1 simple revision"
          ],
          [
            "Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período",
            "Paid traffic for 3 months (90 days) + organic in the same period"
          ]
        ]
      },
      {
        "nome": [
          "Pacote 2",
          "Package 2"
        ],
        "resumo": [
          "2 vídeos verticais",
          "2 vertical videos"
        ],
        "preco": "R$ 550",
        "destaque": true,
        "itens": [
          [
            "2 roteiros humanizados",
            "2 humanized scripts"
          ],
          [
            "2 fotos lifestyle",
            "2 lifestyle photos"
          ],
          [
            "1 revisão simples",
            "1 simple revision"
          ],
          [
            "Edição completa",
            "Full editing"
          ],
          [
            "Legendas",
            "Captions"
          ],
          [
            "Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)",
            "Light visual/sound effects (pacing, transitions, on-screen text)"
          ],
          [
            "Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período",
            "Paid traffic for 3 months (90 days) + organic in the same period"
          ]
        ]
      },
      {
        "nome": [
          "Pacote 3",
          "Package 3"
        ],
        "resumo": [
          "4 vídeos verticais",
          "4 vertical videos"
        ],
        "preco": "R$ 1100",
        "itens": [
          [
            "4 roteiros humanizados (estratégia de campanha)",
            "4 humanized scripts (campaign strategy)"
          ],
          [
            "Edição premium (mais dinâmica + acabamento)",
            "Premium editing (more pacing + finishing)"
          ],
          [
            "2 revisões",
            "2 revisions"
          ],
          [
            "Legendas",
            "Captions"
          ],
          [
            "Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)",
            "Light visual/sound effects (pacing, transitions, on-screen text)"
          ],
          [
            "Tráfego pago por 6 meses (180 dias) + orgânico no mesmo período",
            "Paid traffic for 6 months (180 days) + organic in the same period"
          ]
        ]
      }
    ]
  },
  {
    "titulo": [
      "Investimento Influencer",
      "Influencer Investment"
    ],
    "tag": "Influencer",
    "pacotes": [
      {
        "nome": [
          "Pacote 1",
          "Package 1"
        ],
        "resumo": [
          "Sequência de 3 stories",
          "Sequence of 3 stories"
        ],
        "preco": "R$ 150",
        "itens": [
          [
            "Copy curta",
            "Short copy"
          ],
          [
            "Marcação",
            "Tag/mention"
          ],
          [
            "Link/CTA (se tiver)",
            "Link/CTA (if any)"
          ]
        ]
      },
      {
        "nome": [
          "Pacote 2",
          "Package 2"
        ],
        "resumo": [
          "1 Reels + 1 foto de capa",
          "1 Reels + 1 cover photo"
        ],
        "preco": "R$ 300",
        "itens": [
          [
            "1 foto de capa",
            "1 cover photo"
          ],
          [
            "Sequência de 3 stories com CTA",
            "Sequence of 3 stories with CTA"
          ],
          [
            "Marcação",
            "Tag/mention"
          ]
        ]
      },
      {
        "nome": [
          "Pacote 3",
          "Package 3"
        ],
        "resumo": [
          "2 Reels/mês + 1 story semanal (4 no mês)",
          "2 Reels/month + 1 weekly story (4 a month)"
        ],
        "preco": "R$ 700/mês",
        "itens": [
          [
            "Planejamento do mês (datas) + CTA",
            "Monthly planning (dates) + CTA"
          ],
          [
            "Marcação",
            "Tag/mention"
          ]
        ]
      }
    ]
  }
];

export const PROCESSO: Passo[] = [
  {
    "num": "01",
    "titulo": [
      "Briefing & objetivos",
      "Briefing & goals"
    ],
    "prazo": [
      "~10 min",
      "~10 min"
    ],
    "resumo": [
      "Alinhamos produto, público e meta da campanha.",
      "We align product, audience and campaign goal."
    ],
    "detalhes": [
      [
        "Você envia briefing, referências, tom de voz e infos do produto.",
        "You send the briefing, references, tone of voice and product info."
      ],
      [
        "Definimos objetivo (awareness/conversão), formato (9:16) e CTA.",
        "We set the goal (awareness/conversion), format (9:16) and CTA."
      ],
      [
        "Eu confirmo entregáveis e prazo ideal para gravação e edição.",
        "I confirm deliverables and the ideal timeline for shooting and editing."
      ]
    ]
  },
  {
    "num": "02",
    "titulo": [
      "Proposta & pacote ideal",
      "Proposal & ideal package"
    ],
    "prazo": [
      "rápido",
      "quick"
    ],
    "resumo": [
      "Eu te envio opções de entregáveis com valores e prazos.",
      "I send you deliverable options with prices and timelines."
    ],
    "detalhes": [
      [
        "Pacotes com variações (hook/ângulo), versões com e sem legenda.",
        "Packages with variations (hook/angle), captioned and uncaptioned versions."
      ],
      [
        "Definição de uso (orgânico/ads/whitelisting, se houver).",
        "Usage definition (organic/ads/whitelisting, if any)."
      ],
      [
        "Se quiser: opções extras (UGC foto, cortes, thumbnails).",
        "If you want: extra options (photo UGC, cuts, thumbnails)."
      ]
    ]
  },
  {
    "num": "03",
    "titulo": [
      "Roteiro & aprovação rápida",
      "Script & quick approval"
    ],
    "prazo": [
      "1–2 dias",
      "1–2 days"
    ],
    "resumo": [
      "Eu proponho hooks + estrutura do vídeo pra aprovar antes de gravar.",
      "I propose hooks + video structure for approval before shooting."
    ],
    "detalhes": [
      [
        "Hooks + bullet points (o que precisa ser dito / mostrado).",
        "Hooks + bullet points (what needs to be said / shown)."
      ],
      [
        "Você aprova ou ajusta (sem burocracia).",
        "You approve or adjust (no bureaucracy)."
      ],
      [
        "Depois disso, a gravação segue alinhada e eficiente.",
        "After that, the shoot runs aligned and efficient."
      ]
    ]
  },
  {
    "num": "04",
    "titulo": [
      "Produção & gravação",
      "Production & shooting"
    ],
    "prazo": [
      "1–3 dias",
      "1–3 days"
    ],
    "resumo": [
      "Eu gravo com estética alinhada ao produto + tendência.",
      "I shoot with an aesthetic aligned to the product + trend."
    ],
    "detalhes": [
      [
        "Gravação em luz natural/estúdio (o que combinar mais).",
        "Shooting in natural light/studio (whichever fits best)."
      ],
      [
        "Cenas de uso real + close do produto + prova social (quando faz sentido).",
        "Real-use scenes + product close-ups + social proof (when it makes sense)."
      ],
      [
        "Captação pensando em cortes pra ads e variações.",
        "Footage planned for ad cuts and variations."
      ]
    ]
  },
  {
    "num": "05",
    "titulo": [
      "Edição & entrega final",
      "Editing & final delivery"
    ],
    "prazo": [
      "2–5 dias",
      "2–5 days"
    ],
    "resumo": [
      "Eu edito (dinâmico + legendas) e te entrego tudo organizado.",
      "I edit (dynamic + captions) and deliver everything organized."
    ],
    "detalhes": [
      [
        "Entrega em alta + versão otimizada pra plataformas (Reels/TikTok).",
        "High-res delivery + version optimized for platforms (Reels/TikTok)."
      ],
      [
        "Organização por pasta e nomes fáceis (ex.: Hook1_V1, Hook1_V2...).",
        "Organized by folder with clear names (e.g. Hook1_V1, Hook1_V2...)."
      ],
      [
        "1 rodada de ajustes leves incluída (ex.: texto/legenda/corte).",
        "1 round of light adjustments included (e.g. text/caption/cut)."
      ]
    ]
  }
];

export const METRICAS: Rede[] = [
  {
    "rede": "Instagram",
    "seguidores": "+3.2K",
    "views": "+36.1K",
    "locais": [
      {
        "rotulo": [
          "Brasil",
          "Brazil"
        ],
        "pct": 96.7
      },
      {
        "rotulo": [
          "Curitiba",
          "Curitiba"
        ],
        "pct": 29.5
      }
    ],
    "genero": [
      {
        "rotulo": [
          "Feminino",
          "Female"
        ],
        "pct": 83.9,
        "cor": "#d8b45a"
      },
      {
        "rotulo": [
          "Masculino",
          "Male"
        ],
        "pct": 16.1,
        "cor": "#efe3d0"
      }
    ],
    "idades": [
      {
        "rotulo": [
          "18–24",
          "18–24"
        ],
        "pct": 36.9,
        "cor": "#d8b45a"
      },
      {
        "rotulo": [
          "25–34",
          "25–34"
        ],
        "pct": 47.5,
        "cor": "#efe3d0"
      },
      {
        "rotulo": [
          "+35",
          "+35"
        ],
        "pct": 15.6,
        "cor": "#5f9184"
      }
    ],
    "foto": "/static/img/midiakit/phone-instagram.webp",
    "lado": "dir"
  },
  {
    "rede": "TikTok",
    "seguidores": "+7K",
    "views": "+17K",
    "locais": [
      {
        "rotulo": [
          "Brasil",
          "Brazil"
        ],
        "pct": 99.4
      }
    ],
    "genero": [
      {
        "rotulo": [
          "Feminino",
          "Female"
        ],
        "pct": 78.0,
        "cor": "#d8b45a"
      },
      {
        "rotulo": [
          "Masculino",
          "Male"
        ],
        "pct": 21.0,
        "cor": "#efe3d0"
      }
    ],
    "idades": [
      {
        "rotulo": [
          "18–24",
          "18–24"
        ],
        "pct": 54.2,
        "cor": "#d8b45a"
      },
      {
        "rotulo": [
          "25–34",
          "25–34"
        ],
        "pct": 35.9,
        "cor": "#efe3d0"
      },
      {
        "rotulo": [
          "+35",
          "+35"
        ],
        "pct": 9.9,
        "cor": "#5f9184"
      }
    ],
    "foto": "/static/img/midiakit/phone-tiktok.webp",
    "lado": "esq"
  }
];

export const VIDEOS: Categoria[] = [
  {
    "slug": "beleza",
    "nome": "Beleza",
    "videos": [
      {
        "titulo": "L'Oréal Paris",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Loreal.webm",
        "poster": "/static/img/midiakit/videos/beleza-loreal"
      },
      {
        "titulo": "Dove",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Dove.webm",
        "poster": "/static/img/midiakit/videos/beleza-dove"
      },
      {
        "titulo": "Pantene",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Pantene.webm",
        "poster": "/static/img/midiakit/videos/beleza-pantene"
      },
      {
        "titulo": "Maquiagem",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Maquiagem.webm",
        "poster": "/static/img/midiakit/videos/beleza-maquiagem"
      },
      {
        "titulo": "Elseve",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Elseve.webm",
        "poster": "/static/img/midiakit/videos/beleza-elseve"
      }
    ]
  },
  {
    "slug": "eventos",
    "nome": "Eventos",
    "videos": [
      {
        "titulo": "Natura Yoga",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Natura_Yoga.webm",
        "poster": "/static/img/midiakit/videos/eventos-natura-yoga"
      },
      {
        "titulo": "Mercado Livre",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Mercado_Livre.webm",
        "poster": "/static/img/midiakit/videos/eventos-mercado-livre"
      },
      {
        "titulo": "Top Girls Experience",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Top_Girls_Experience.webm",
        "poster": "/static/img/midiakit/videos/eventos-top-girls-experience"
      },
      {
        "titulo": "Roda das Deusas",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Roda_Das_Deusas.webm",
        "poster": "/static/img/midiakit/videos/eventos-roda-das-deusas"
      }
    ]
  },
  {
    "slug": "fitness",
    "nome": "Fitness",
    "videos": [
      {
        "titulo": "Probiótica",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/fitness/Probiotica.webm",
        "poster": "/static/img/midiakit/videos/fitness-probiotica"
      },
      {
        "titulo": "Crossfit",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/fitness/Crossfit.webm",
        "poster": "/static/img/midiakit/videos/fitness-crossfit"
      }
    ]
  },
  {
    "slug": "gastronomia",
    "nome": "Gastronomia",
    "videos": [
      {
        "titulo": "Café Onírico",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/gastronomia/Cafe_Onirico.webm",
        "poster": "/static/img/midiakit/videos/gastronomia-cafe-onirico"
      },
      {
        "titulo": "Doces Bliss",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/gastronomia/Doces_Bliss.webm",
        "poster": "/static/img/midiakit/videos/gastronomia-doces-bliss"
      }
    ]
  },
  {
    "slug": "moda",
    "nome": "Moda",
    "videos": [
      {
        "titulo": "Pé de Algodão",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Pe_De_Algodao.webm",
        "poster": "/static/img/midiakit/videos/moda-pe-de-algodao"
      },
      {
        "titulo": "Pri Joias",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Pri_Joias.webm",
        "poster": "/static/img/midiakit/videos/moda-pri-joias"
      },
      {
        "titulo": "Atelie Cris Tavares",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Atelie_Cris_Tavares.webm",
        "poster": "/static/img/midiakit/videos/moda-atelie-cris-tavares"
      }
    ]
  },
  {
    "slug": "saude",
    "nome": "Saúde",
    "videos": [
      {
        "titulo": "Redoxon",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/saude/Redoxon.webm",
        "poster": "/static/img/midiakit/videos/saude-redoxon"
      },
      {
        "titulo": "Gino-canisten",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/saude/Ginocanisten.webm",
        "poster": "/static/img/midiakit/videos/saude-ginocanisten"
      }
    ]
  },
  {
    "slug": "parcerias",
    "nome": "Parcerias",
    "videos": [
      {
        "titulo": "Comu Delas",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/parcerias/Comu_Delas.webm",
        "poster": "/static/img/midiakit/videos/parcerias-comu-delas"
      }
    ]
  },
  {
    "slug": "diario",
    "nome": "Diário de uma fotógrafa",
    "videos": [
      {
        "titulo": "Isaresponde",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Isaresponde.webm",
        "poster": "/static/img/midiakit/videos/diario-isaresponde"
      },
      {
        "titulo": "Bate Papo",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Bate_Papo.webm",
        "poster": "/static/img/midiakit/videos/diario-bate-papo"
      },
      {
        "titulo": "Musa Day Afrodite Lançamento",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Musa%20Day%20Afrodite%20Lan%C3%A7amento.webm",
        "poster": "/static/img/midiakit/videos/diario-musa-day-afrodite-lancamento"
      },
      {
        "titulo": "Musaday",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Musaday.webm",
        "poster": "/static/img/midiakit/videos/diario-musaday"
      },
      {
        "titulo": "Projeto Metamorfose",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Projeto_Metamorfose.webm",
        "poster": "/static/img/midiakit/videos/diario-projeto-metamorfose"
      },
      {
        "titulo": "Organizando acessorios",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Organizando_acessorios.webm",
        "poster": "/static/img/midiakit/videos/diario-organizando-acessorios"
      },
      {
        "titulo": "Dicas de Camera",
        "src": "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Dicas_de_Camera.webm",
        "poster": "/static/img/midiakit/videos/diario-dicas-de-camera"
      }
    ]
  }
];
