export function Lightbox() {
  return (
    <dialog id="lightbox" className="lightbox" aria-label="Visualizador de fotos">
      <button className="lightbox__close" data-lb-close aria-label="Fechar">
        ✕
      </button>
      <button className="lightbox__nav lightbox__nav--prev" data-lb-prev aria-label="Anterior">
        ‹
      </button>
      <figure className="lightbox__fig">
        <img data-lb-img alt="" />
        <figcaption className="lightbox__cap">
          <span data-lb-tema></span>
          <span className="lightbox__count" data-lb-count></span>
        </figcaption>
      </figure>
      <button className="lightbox__nav lightbox__nav--next" data-lb-next aria-label="Próxima">
        ›
      </button>
    </dialog>
  );
}
