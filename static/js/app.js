// JS do site. ~vanilla, substitui Radix + Embla + react-masonry do React.
// Carregado com defer; o HTMX cuida do filtro da galeria por conta propria.

(() => {
	"use strict";

	// ── Fade-in dos thumbs ───────────────────────────────────────────────────────
	// O thumb sobe sobre o LQIP quando carrega. Trata o caso cacheado (o browser
	// ja tem a imagem -> `load` nao dispara): se `complete`, revela na hora.
	const revealGrid = () => {
		document
			.querySelectorAll(".masonry__img:not(.is-loaded)")
			.forEach((img) => {
				if (img.complete) img.classList.add("is-loaded");
				else
					img.addEventListener("load", () => img.classList.add("is-loaded"), {
						once: true,
					});
			});
	};
	revealGrid();
	document.body.addEventListener("htmx:afterSwap", revealGrid);

	// ── Lightbox ───────────────────────────────────────────────────────────────
	// <dialog> nativo: Esc, focus trap e ::backdrop de graca -- o React fazia isso
	// a mao e sem focus trap (o Tab escapava para a pagina atras). As imagens ja
	// estao no DOM (o grid inteiro foi renderizado), entao NAO ha hx-get aqui:
	// buscar no servidor um dado que esta a 3 linhas seria HTMX por ideologia.
	const dlg = document.getElementById("lightbox");
	if (dlg) {
		const imgEl = dlg.querySelector("[data-lb-img]");
		const temaEl = dlg.querySelector("[data-lb-tema]");
		const contEl = dlg.querySelector("[data-lb-count]");
		let itens = [];
		let i = 0;

		const refresh = () => {
			itens = [...document.querySelectorAll(".masonry [data-idx]")];
		};

		const preload = (el) => {
			if (!el) return;
			const img = new Image();
			img.src = el.dataset.full;
		};

		const mostrar = (n) => {
			if (itens.length === 0) return;
			i = (n + itens.length) % itens.length;
			const el = itens[i];
			imgEl.src = el.dataset.full;
			imgEl.alt = el.dataset.alt || "";
			if (temaEl) temaEl.textContent = el.dataset.temaLabel || "";
			if (contEl) contEl.textContent = `${i + 1} / ${itens.length}`;
			// vizinhos: martelar as setas fica instantaneo, sem piscar branco
			preload(itens[(i + 1) % itens.length]);
			preload(itens[(i - 1 + itens.length) % itens.length]);
		};

		// Guard: abrir o <dialog> dentro de um handler de clique faz o MESMO clique
		// vazar para o backdrop e fechar na hora. Enquanto justOpened, o handler de
		// backdrop ignora o clique. Limpo no proximo frame.
		let justOpened = false;

		// clique num item do grid abre o lightbox
		document.addEventListener("click", (e) => {
			const alvo = e.target.closest(".masonry [data-idx]");
			if (!alvo) return;
			refresh();
			mostrar(itens.indexOf(alvo));
			dlg.showModal();
			justOpened = true;
			requestAnimationFrame(() => {
				justOpened = false;
			});
		});

		// navegacao por teclado
		dlg.addEventListener("keydown", (e) => {
			if (e.key === "ArrowRight") {
				e.preventDefault();
				mostrar(i + 1);
			} else if (e.key === "ArrowLeft") {
				e.preventDefault();
				mostrar(i - 1);
			}
		});

		// clique no backdrop (fora da imagem) fecha -- exceto o clique que acabou de
		// abrir (ver justOpened).
		dlg.addEventListener("click", (e) => {
			if (justOpened) return;
			if (e.target === dlg) dlg.close();
		});

		// botoes prev/next/close
		dlg
			.querySelectorAll("[data-lb-prev]")
			.forEach((b) => b.addEventListener("click", () => mostrar(i - 1)));
		dlg
			.querySelectorAll("[data-lb-next]")
			.forEach((b) => b.addEventListener("click", () => mostrar(i + 1)));
		dlg
			.querySelectorAll("[data-lb-close]")
			.forEach((b) => b.addEventListener("click", () => dlg.close()));

		// depois de um swap do HTMX (troca de filtro), o grid mudou -> reindexa.
		// Unica armadilha de integracao HTMX<->JS do projeto.
		document.body.addEventListener("htmx:afterSwap", refresh);
	}

	// ── Carrossel (/sobre) ───────────────────────────────────────────────────────
	// scroll-snap CSS faz o arrasto/touch; as setas so empurram um item. Substitui
	// o Embla. Perde o loop infinito -- aceitavel para 8 fotos.
	document.querySelectorAll(".carrossel-wrap").forEach((wrap) => {
		const track = wrap.querySelector("[data-carrossel]");
		if (!track) return;
		const passo = () =>
			track.querySelector(".carrossel__item")?.offsetWidth ||
			track.clientWidth * 0.8;
		wrap
			.querySelector("[data-carrossel-prev]")
			?.addEventListener("click", () =>
				track.scrollBy({ left: -passo(), behavior: "smooth" }),
			);
		wrap
			.querySelector("[data-carrossel-next]")
			?.addEventListener("click", () =>
				track.scrollBy({ left: passo(), behavior: "smooth" }),
			);
	});

	// ── Menu mobile ──────────────────────────────────────────────────────────────
	// Mesmo padrao do lightbox: <dialog> + guard justOpened contra o clique que
	// abre vazar para o backdrop. Substitui o Radix Sheet do React.
	const menu = document.getElementById("menu");
	if (menu) {
		let menuJustOpened = false;
		document.querySelectorAll("[data-menu-abrir]").forEach((b) =>
			b.addEventListener("click", () => {
				menu.showModal();
				menuJustOpened = true;
				requestAnimationFrame(() => {
					menuJustOpened = false;
				});
			}),
		);
		menu
			.querySelectorAll("[data-menu-fechar], .menu__links a")
			.forEach((el) => el.addEventListener("click", () => menu.close()));
		menu.addEventListener("click", (e) => {
			if (menuJustOpened) return;
			if (e.target === menu) menu.close();
		});
	}

	// ── Navbar hide-on-scroll ────────────────────────────────────────────────────
	// Some ao descer, volta ao subir. No React essa logica estava DUPLICADA em
	// NavBar.tsx e PhotoGallery.tsx; aqui e uma classe no <body>, uma vez so.
	let ultimoY = window.scrollY;
	window.addEventListener(
		"scroll",
		() => {
			const y = window.scrollY;
			const d = y - ultimoY;
			if (y < 25) {
				document.body.classList.remove("nav-oculta");
				ultimoY = y;
				return;
			}
			if (Math.abs(d) < 8) return; // ignora micro-scroll (tremida)
			document.body.classList.toggle("nav-oculta", d > 0); // desceu -> esconde
			ultimoY = y;
		},
		{ passive: true },
	);
})();
