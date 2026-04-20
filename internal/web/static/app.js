(function () {
  const world = document.getElementById('world');
  const viewport = document.getElementById('viewport');
  const slider = document.getElementById('zoom');
  const label = document.getElementById('zoomLabel');

  const baseW = Number(world.dataset.width);
  const baseH = Number(world.dataset.height);
  const tileW = Number(world.dataset.tileW);
  const tileH = Number(world.dataset.tileH);

  // Apply fixed dimensions to every tile image once.
  for (const img of world.getElementsByTagName('img')) {
    img.style.width = tileW + 'px';
    img.style.height = tileH + 'px';
  }

  world.style.width = baseW + 'px';
  world.style.height = baseH + 'px';

  function applyScale(scale) {
    world.style.transform = 'scale(' + scale + ')';
    world.style.width = (baseW * scale) + 'px';
    world.style.height = (baseH * scale) + 'px';
    label.textContent = Math.round(scale * 100) + '%';
  }

  // Auto-fit the initial zoom so the map fits the viewport with some breathing room.
  const fit = Math.min(
    (viewport.clientWidth - 120) / baseW,
    (viewport.clientHeight - 200) / baseH,
    1.5
  );
  const initial = Math.max(0.1, Math.min(1.5, fit));
  slider.value = Math.round(initial * 100);
  applyScale(initial);

  slider.addEventListener('input', () => applyScale(slider.value / 100));
})();
