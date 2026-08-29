(function () {
  var STORAGE_KEY = "golol.builds";
  var ITEM_SLOTS = 7;

  var page = document.querySelector(".builds-page");
  if (!page) return;

  var slotsAttr = parseInt(page.getAttribute("data-item-slots"), 10);
  if (slotsAttr > 0) ITEM_SLOTS = slotsAttr;

  var list = document.getElementById("builds-list");
  var mode = document.getElementById("builds-mode");
  var addBtn = document.getElementById("add-build");
  var exportBtn = document.getElementById("export-builds");
  var importBtn = document.getElementById("import-builds");
  var importPanel = document.getElementById("import-panel");
  var importJson = document.getElementById("import-json");
  var importConfirm = document.getElementById("import-confirm");
  var importCancel = document.getElementById("import-cancel");
  var champPicker = document.getElementById("champ-picker");
  var itemPicker = document.getElementById("item-picker");
  var champSearch = document.getElementById("champ-search");
  var itemSearch = document.getElementById("item-search");
  var champCount = document.getElementById("champ-count");
  var itemCount = document.getElementById("item-count");
  var champEmpty = document.getElementById("champ-empty");
  var itemEmpty = document.getElementById("item-empty");
  var champTotal = champPicker ? champPicker.querySelectorAll("[data-kind=champion]").length : 0;
  var itemTotal = itemPicker ? itemPicker.querySelectorAll("[data-kind=item]").length : 0;

  var builds = load();
  var selectedId = null;
  var bonusIndex = loadBonusIndex();
  var flyout = ensureFlyout();
  var flyoutCache = {};
  var flyoutAnchor = null;
  var flyoutItemId = "";
  var flyoutTimer = null;
  var flyoutAbort = null;

  if (addBtn) {
    addBtn.addEventListener("click", function () {
      var build = { id: newId(), champion: null, items: [] };
      builds.push(build);
      selectedId = build.id;
      persist();
      render();
    });
  }

  if (exportBtn) {
    exportBtn.addEventListener("click", exportBuilds);
  }

  if (importBtn) {
    importBtn.addEventListener("click", function () {
      openImportPanel();
    });
  }

  if (importConfirm) {
    importConfirm.addEventListener("click", function () {
      applyImport(importJson ? importJson.value : "");
    });
  }

  if (importCancel) {
    importCancel.addEventListener("click", closeImportPanel);
  }

  if (importJson) {
    importJson.addEventListener("paste", function () {
      var el = importJson;
      setTimeout(function () {
        applyImport(el.value);
      }, 0);
    });
  }

  if (list) {
    list.addEventListener("click", function (e) {
      var remove = e.target.closest("[data-remove-build]");
      if (remove) {
        e.preventDefault();
        removeBuild(remove.getAttribute("data-remove-build"));
        return;
      }
      var itemBtn = e.target.closest("[data-remove-item]");
      if (itemBtn) {
        e.preventDefault();
        var article = itemBtn.closest("[data-build-id]");
        if (!article) return;
        var idx = parseInt(itemBtn.getAttribute("data-remove-item"), 10);
        removeItem(article.getAttribute("data-build-id"), idx);
        return;
      }
      var article = e.target.closest("[data-build-id]");
      if (!article) return;
      selectBuild(article.getAttribute("data-build-id"));
    });

    list.addEventListener("keydown", function (e) {
      if (e.key !== "Enter" && e.key !== " ") return;
      if (e.target.closest("button")) return;
      var article = e.target.closest("[data-build-id]");
      if (!article) return;
      e.preventDefault();
      selectBuild(article.getAttribute("data-build-id"));
    });

    list.addEventListener("pointerover", function (e) {
      if (e.pointerType === "touch") return;
      var btn = e.target.closest("[data-item-id]");
      if (!btn || !list.contains(btn)) return;
      if (e.relatedTarget && btn.contains(e.relatedTarget)) return;
      requestShowFlyout(btn);
    });

    list.addEventListener("pointerout", function (e) {
      if (e.pointerType === "touch") return;
      var btn = e.target.closest("[data-item-id]");
      if (!btn || !list.contains(btn)) return;
      if (e.relatedTarget && btn.contains(e.relatedTarget)) return;
      requestHideFlyout();
    });

    list.addEventListener("focusin", function (e) {
      var btn = e.target.closest("[data-item-id]");
      if (btn && list.contains(btn)) requestShowFlyout(btn);
    });

    list.addEventListener("focusout", function (e) {
      var btn = e.target.closest("[data-item-id]");
      if (!btn) return;
      if (e.relatedTarget && (btn.contains(e.relatedTarget) || flyout.contains(e.relatedTarget))) return;
      requestHideFlyout();
    });
  }

  if (flyout) {
    flyout.addEventListener("pointerenter", cancelFlyoutTimer);
    flyout.addEventListener("pointerleave", requestHideFlyout);
    flyout.addEventListener("focusin", cancelFlyoutTimer);
    flyout.addEventListener("focusout", function (e) {
      if (e.relatedTarget && flyout.contains(e.relatedTarget)) return;
      requestHideFlyout();
    });
    flyout.addEventListener("click", function (e) {
      e.stopPropagation();
      var mini = e.target.closest("[data-flyout-item]");
      if (!mini) return;
      e.preventDefault();
      openFlyout(flyoutAnchor, mini.getAttribute("data-flyout-item"));
    });
  }

  document.addEventListener("keydown", function (e) {
    if (e.key !== "Escape") return;
    hideFlyout();
    closeImportPanel();
  });

  window.addEventListener("scroll", repositionFlyout, true);
  window.addEventListener("resize", repositionFlyout);

  if (champPicker) {
    champPicker.addEventListener("click", function (e) {
      var btn = e.target.closest("[data-kind=champion]");
      if (!btn) return;
      addChampion(pieceFrom(btn));
    });
  }

  if (itemPicker) {
    itemPicker.addEventListener("click", function (e) {
      var btn = e.target.closest("[data-kind=item]");
      if (!btn) return;
      addItem(pieceFrom(btn));
    });
  }

  if (champSearch) {
    champSearch.addEventListener("input", function () {
      filterPicker(champPicker, champSearch.value, champCount, champEmpty, champTotal, "campeón", "campeones");
    });
  }

  if (itemSearch) {
    itemSearch.addEventListener("input", function () {
      filterPicker(itemPicker, itemSearch.value, itemCount, itemEmpty, itemTotal, "objeto", "objetos");
    });
  }

  render();

  function selectBuild(id) {
    selectedId = id;
    render();
  }

  function addChampion(piece) {
    var build = selectedBuild();
    if (!build) {
      flashMode();
      return;
    }
    if (!piece.id) return;
    build.champion = piece;
    persist();
    render();
  }

  function addItem(piece) {
    var build = selectedBuild();
    if (!build) {
      flashMode();
      return;
    }
    if (!piece.id) return;
    if (build.items.some(function (it) { return it.id === piece.id; })) return;
    if (build.items.length >= ITEM_SLOTS) {
      setMode("Esta build ya tiene " + ITEM_SLOTS + " objetos.");
      flashMode();
      return;
    }
    build.items.push(piece);
    persist();
    render();
  }

  function removeItem(buildId, index) {
    var build = findBuild(buildId);
    if (!build) return;
    if (index < 0 || index >= build.items.length) return;
    build.items.splice(index, 1);
    selectedId = buildId;
    persist();
    render();
  }

  function removeBuild(id) {
    builds = builds.filter(function (b) { return b.id !== id; });
    if (selectedId === id) selectedId = null;
    persist();
    render();
  }

  function selectedBuild() {
    return findBuild(selectedId);
  }

  function findBuild(id) {
    if (!id) return null;
    for (var i = 0; i < builds.length; i++) {
      if (builds[i].id === id) return builds[i];
    }
    return null;
  }

  function exportBuilds() {
    var text = JSON.stringify(builds, null, 2);
    copyText(text).then(function () {
      setMode("Colecciones copiadas al portapapeles.");
      flashMode();
    }).catch(function () {
      openImportPanel(text);
      setMode("No se pudo copiar. El JSON está en el recuadro para que lo copies a mano.");
      flashMode();
    });
  }

  function openImportPanel(prefill) {
    if (!importPanel) return;
    importPanel.hidden = false;
    if (importJson) {
      importJson.value = typeof prefill === "string" ? prefill : "";
      importJson.focus();
      importJson.select();
    }
    if (typeof prefill !== "string") {
      setMode("Pega un JSON válido de colecciones.");
    }
  }

  function closeImportPanel() {
    if (!importPanel || importPanel.hidden) return;
    importPanel.hidden = true;
    if (importJson) importJson.value = "";
  }

  function applyImport(raw) {
    var incoming;
    try {
      incoming = parseImported(raw);
    } catch (err) {
      setMode(err.message || "Ese texto no es un JSON válido.");
      flashMode();
      return;
    }
    builds = incoming;
    selectedId = null;
    persist();
    closeImportPanel();
    render();
    var n = builds.length;
    setMode(n === 1 ? "Se importó 1 colección." : "Se importaron " + n + " colecciones.");
    flashMode();
  }

  function parseImported(text) {
    var data;
    try {
      data = JSON.parse(String(text || "").trim());
    } catch (err) {
      throw new Error("Ese texto no es un JSON válido.");
    }
    var list;
    if (Array.isArray(data)) {
      list = data;
    } else if (data && typeof data === "object") {
      list = [data];
    } else {
      throw new Error("El JSON no contiene colecciones.");
    }
    var out = [];
    var seen = {};
    for (var i = 0; i < list.length; i++) {
      var b = normalizeBuild(list[i]);
      if (!b) continue;
      if (seen[b.id]) b.id = newId();
      seen[b.id] = true;
      out.push(b);
    }
    if (list.length > 0 && out.length === 0) {
      throw new Error("El JSON no contiene colecciones.");
    }
    return out;
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text);
    }
    return new Promise(function (resolve, reject) {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      try {
        if (document.execCommand("copy")) resolve();
        else reject(new Error("copy"));
      } catch (err) {
        reject(err);
      } finally {
        document.body.removeChild(ta);
      }
    });
  }

  function persist() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(builds));
    } catch (err) {
      setMode("No se pudo guardar en el almacenamiento del navegador.");
    }
  }

  function load() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      var data = JSON.parse(raw);
      if (!Array.isArray(data)) return [];
      return data.map(normalizeBuild).filter(Boolean);
    } catch (err) {
      return [];
    }
  }

  function normalizeBuild(raw) {
    if (!raw || typeof raw !== "object") return null;
    var id = String(raw.id || "");
    if (!id) id = newId();
    return {
      id: id,
      champion: normalizePiece(raw.champion),
      items: normalizeItems(raw.items)
    };
  }

  function normalizeItems(raw) {
    if (!Array.isArray(raw)) return [];
    var out = [];
    for (var i = 0; i < raw.length && out.length < ITEM_SLOTS; i++) {
      var piece = normalizePiece(raw[i]);
      if (piece) out.push(piece);
    }
    return out;
  }

  function normalizePiece(raw) {
    if (!raw || typeof raw !== "object" || !raw.id) return null;
    return {
      id: String(raw.id),
      name: String(raw.name || ""),
      icon: String(raw.icon || "")
    };
  }

  function pieceFrom(btn) {
    return {
      id: btn.getAttribute("data-id") || "",
      name: btn.getAttribute("data-name") || "",
      icon: btn.getAttribute("data-icon") || ""
    };
  }

  function newId() {
    if (window.crypto && crypto.randomUUID) return crypto.randomUUID();
    return "b-" + Date.now().toString(36) + "-" + Math.random().toString(36).slice(2, 8);
  }

  function render() {
    renderList();
    renderMode();
    syncPickerHighlights();
  }

  function renderList() {
    hideFlyout();
    if (!list) return;
    list.replaceChildren();
    if (builds.length === 0) {
      var empty = document.createElement("p");
      empty.className = "empty";
      empty.textContent = "Aún no hay builds. Pulsa Añadir build para crear una.";
      list.appendChild(empty);
      return;
    }
    builds.forEach(function (build) {
      list.appendChild(renderBuild(build));
    });
  }

  function renderBuild(build) {
    var selected = build.id === selectedId;
    var article = document.createElement("article");
    article.className = "build" + (selected ? " is-on" : "");
    article.setAttribute("data-build-id", build.id);
    article.setAttribute("tabindex", "0");
    article.setAttribute("aria-current", selected ? "true" : "false");
    article.setAttribute("aria-label", buildLabel(build, selected));

    article.appendChild(renderChampSlot(build.champion));
    article.appendChild(renderItemSlots(build));

    var remove = document.createElement("button");
    remove.type = "button";
    remove.className = "build-remove";
    remove.setAttribute("data-remove-build", build.id);
    remove.setAttribute("aria-label", "Quitar build");
    remove.textContent = "Quitar";
    article.appendChild(remove);

    var totals = renderBonuses(build);
    if (totals) article.appendChild(totals);

    return article;
  }

  function renderChampSlot(champ) {
    var wrap = document.createElement("div");
    wrap.className = "build-champ" + (champ ? "" : " is-empty");
    var slot = document.createElement("div");
    slot.className = "slot slot-champ" + (champ ? "" : " is-empty");
    if (champ) {
      var img = document.createElement("img");
      img.src = champ.icon;
      img.alt = champ.name || "";
      img.width = 64;
      img.height = 64;
      slot.appendChild(img);
    }
    wrap.appendChild(slot);
    var name = document.createElement("span");
    name.className = "slot-name";
    name.textContent = champ ? (champ.name || champ.id) : "Campeón";
    wrap.appendChild(name);
    return wrap;
  }

  function renderItemSlots(build) {
    var ol = document.createElement("ol");
    ol.className = "build-items";
    for (var i = 0; i < ITEM_SLOTS; i++) {
      var li = document.createElement("li");
      var piece = build.items[i];
      if (piece) {
        var btn = document.createElement("button");
        btn.type = "button";
        btn.className = "slot";
        btn.setAttribute("data-remove-item", String(i));
        btn.setAttribute("data-item-id", piece.id);
        btn.setAttribute("aria-label", "Quitar " + (piece.name || piece.id));
        var img = document.createElement("img");
        img.src = piece.icon;
        img.alt = piece.name || "";
        img.width = 48;
        img.height = 48;
        btn.appendChild(img);
        li.appendChild(btn);
      } else {
        var empty = document.createElement("div");
        empty.className = "slot is-empty";
        empty.setAttribute("aria-hidden", "true");
        li.appendChild(empty);
      }
      ol.appendChild(li);
    }
    return ol;
  }

  function ensureFlyout() {
    var el = document.getElementById("item-flyout");
    if (el) return el;
    el = document.createElement("div");
    el.id = "item-flyout";
    el.className = "item-flyout";
    el.hidden = true;
    el.setAttribute("role", "tooltip");
    document.body.appendChild(el);
    return el;
  }

  function requestShowFlyout(btn) {
    cancelFlyoutTimer();
    if (!flyout.hidden && flyoutAnchor === btn) return;
    if (!flyout.hidden) {
      openFlyout(btn);
      return;
    }
    flyoutTimer = setTimeout(function () {
      flyoutTimer = null;
      openFlyout(btn);
    }, 80);
  }

  function requestHideFlyout() {
    cancelFlyoutTimer();
    flyoutTimer = setTimeout(function () {
      flyoutTimer = null;
      hideFlyout();
    }, 180);
  }

  function cancelFlyoutTimer() {
    if (flyoutTimer) {
      clearTimeout(flyoutTimer);
      flyoutTimer = null;
    }
  }

  function openFlyout(anchor, id) {
    if (!flyout) return;
    var itemId = id || (anchor && anchor.getAttribute("data-item-id")) || "";
    if (!itemId) return;
    flyoutAnchor = anchor && anchor.isConnected ? anchor : flyoutAnchor;
    if (flyoutItemId === itemId && !flyout.hidden && flyout.innerHTML) {
      placeFlyout(flyoutAnchor);
      return;
    }
    flyoutItemId = itemId;
    if (flyoutCache[itemId]) {
      setFlyoutHTML(flyoutCache[itemId]);
      placeFlyout(flyoutAnchor);
      return;
    }
    setFlyoutMessage("Cargando ficha…");
    placeFlyout(flyoutAnchor);
    if (flyoutAbort) flyoutAbort.abort();
    if (typeof AbortController !== "function") {
      fetchItemDetail(itemId, null);
      return;
    }
    flyoutAbort = new AbortController();
    fetchItemDetail(itemId, flyoutAbort.signal);
  }

  function fetchItemDetail(itemId, signal) {
    var opts = { headers: { Accept: "text/html" } };
    if (signal) opts.signal = signal;
    fetch("/items/" + encodeURIComponent(itemId), opts)
      .then(function (res) {
        if (!res.ok) throw new Error("missing");
        return res.text();
      })
      .then(function (html) {
        flyoutCache[itemId] = html;
        if (flyoutItemId !== itemId) return;
        setFlyoutHTML(html);
        placeFlyout(flyoutAnchor);
      })
      .catch(function (err) {
        if (err && err.name === "AbortError") return;
        if (flyoutItemId !== itemId) return;
        setFlyoutMessage("No se pudo cargar la ficha.");
        placeFlyout(flyoutAnchor);
      });
  }

  function setFlyoutHTML(html) {
    flyout.innerHTML = html;
    flyout.querySelectorAll("[hx-get]").forEach(function (el) {
      var url = el.getAttribute("hx-get") || "";
      var id = url.replace(/^\/items\//, "");
      el.removeAttribute("hx-get");
      el.removeAttribute("hx-target");
      el.removeAttribute("hx-swap");
      el.removeAttribute("hx-push-url");
      if (id) el.setAttribute("data-flyout-item", id);
    });
  }

  function setFlyoutMessage(text) {
    flyout.replaceChildren();
    var p = document.createElement("p");
    p.className = "detail-placeholder";
    p.textContent = text;
    flyout.appendChild(p);
  }

  function placeFlyout(anchor) {
    if (!flyout || !anchor || !anchor.isConnected) return;
    flyout.hidden = false;
    var r = anchor.getBoundingClientRect();
    var w = flyout.offsetWidth;
    var h = flyout.offsetHeight;
    var left = r.right + 10;
    if (left + w > window.innerWidth - 8) {
      left = r.left - w - 10;
      if (left < 8) left = 8;
    }
    var top = r.top;
    if (top + h > window.innerHeight - 8) top = window.innerHeight - h - 8;
    if (top < 8) top = 8;
    flyout.style.left = left + "px";
    flyout.style.top = top + "px";
    describeFlyout(anchor);
  }

  function repositionFlyout() {
    if (!flyout || flyout.hidden) return;
    if (!flyoutAnchor || !flyoutAnchor.isConnected) {
      hideFlyout();
      return;
    }
    placeFlyout(flyoutAnchor);
  }

  function hideFlyout() {
    cancelFlyoutTimer();
    if (flyoutAbort) {
      flyoutAbort.abort();
      flyoutAbort = null;
    }
    flyoutItemId = "";
    flyoutAnchor = null;
    describeFlyout(null);
    if (!flyout) return;
    flyout.hidden = true;
    flyout.replaceChildren();
    flyout.style.left = "";
    flyout.style.top = "";
  }

  function describeFlyout(anchor) {
    document.querySelectorAll("[data-item-id][aria-describedby='item-flyout']").forEach(function (el) {
      el.removeAttribute("aria-describedby");
    });
    if (anchor) anchor.setAttribute("aria-describedby", "item-flyout");
  }

  function renderBonuses(build) {
    var summed = sumBonuses(build.items);
    if (summed.length === 0) return null;
    var wrap = document.createElement("div");
    wrap.className = "build-totals";
    var heading = document.createElement("h3");
    heading.className = "build-totals-label";
    heading.textContent = "Mejoras";
    wrap.appendChild(heading);
    var ul = document.createElement("ul");
    ul.className = "build-bonuses";
    summed.forEach(function (b) {
      var li = document.createElement("li");
      li.className = "build-bonus";
      var amt = document.createElement("span");
      amt.className = "bonus-amt";
      amt.textContent = formatBonusAmount(b.amount, b.percent);
      li.appendChild(amt);
      li.appendChild(document.createTextNode(" " + displayBonusName(b.name)));
      ul.appendChild(li);
    });
    wrap.appendChild(ul);
    return wrap;
  }

  function sumBonuses(pieces) {
    var acc = {};
    (pieces || []).forEach(function (it) {
      if (!it || !it.id) return;
      var list = bonusIndex[it.id];
      if (!list || !list.length) return;
      list.forEach(function (b) {
        var name = String(b.name || "");
        var key = (b.percent ? "1:" : "0:") + name;
        if (!acc[key]) {
          acc[key] = {
            amount: 0,
            percent: !!b.percent,
            name: name,
            rank: typeof b.rank === "number" ? b.rank : 999
          };
        }
        acc[key].amount += Number(b.amount) || 0;
      });
    });
    var out = [];
    Object.keys(acc).forEach(function (k) {
      if (acc[k].amount !== 0) out.push(acc[k]);
    });
    out.sort(function (a, b) {
      if (a.rank !== b.rank) return a.rank - b.rank;
      if (a.percent !== b.percent) return a.percent ? 1 : -1;
      if (a.name < b.name) return -1;
      if (a.name > b.name) return 1;
      return 0;
    });
    return out;
  }

  function formatBonusAmount(n, percent) {
    var v = Number(n) || 0;
    var s;
    if (Math.abs(v - Math.round(v)) < 1e-6) s = String(Math.round(v));
    else s = String(Math.round(v * 100) / 100);
    var sign = v < 0 ? "" : "+";
    return sign + s + (percent ? "%" : "");
  }

  function displayBonusName(name) {
    return String(name || "").replace(/^de\s+/i, "").replace(/^of\s+/i, "");
  }

  function loadBonusIndex() {
    var el = document.getElementById("item-bonuses");
    if (!el) return {};
    try {
      var data = JSON.parse(el.textContent || "{}");
      if (!data || typeof data !== "object" || Array.isArray(data)) return {};
      return data;
    } catch (err) {
      return {};
    }
  }

  function buildLabel(build, selected) {
    var name = build.champion && build.champion.name ? build.champion.name : "sin campeón";
    var n = build.items.length;
    var prefix = selected ? "Build en edición, " : "Build, ";
    var label = prefix + name + ", " + n + (n === 1 ? " objeto" : " objetos");
    var summed = sumBonuses(build.items);
    if (summed.length === 0) return label;
    var parts = summed.map(function (b) {
      return formatBonusAmount(b.amount, b.percent) + " " + displayBonusName(b.name);
    });
    return label + ", " + parts.join(", ");
  }

  function renderMode() {
    var build = selectedBuild();
    if (build) {
      var name = build.champion && build.champion.name ? build.champion.name : "esta build";
      setMode("Editando " + name + ". Elige un campeón o un objeto para añadirlo. Pulsa un objeto de la colección para quitarlo.");
      page.classList.add("is-editing");
    } else {
      setMode("Selecciona una build para editarla.");
      page.classList.remove("is-editing");
    }
  }

  function setMode(text) {
    if (mode) mode.textContent = text;
  }

  function flashMode() {
    if (!mode) return;
    mode.classList.remove("is-flash");
    void mode.offsetWidth;
    mode.classList.add("is-flash");
  }

  function syncPickerHighlights() {
    var build = selectedBuild();
    var champId = build && build.champion ? build.champion.id : "";
    document.querySelectorAll("[data-kind=champion]").forEach(function (el) {
      el.classList.toggle("is-on", champId !== "" && el.getAttribute("data-id") === champId);
    });
    var itemIds = {};
    if (build) {
      build.items.forEach(function (it) { itemIds[it.id] = true; });
    }
    document.querySelectorAll("[data-kind=item]").forEach(function (el) {
      el.classList.toggle("is-on", !!itemIds[el.getAttribute("data-id")]);
    });
  }

  function filterPicker(root, query, countEl, emptyEl, total, singular, plural) {
    if (!root) return;
    var q = fold(query);
    var visible = 0;
    root.querySelectorAll("[data-kind]").forEach(function (el) {
      var match = q === "" || fold(el.getAttribute("data-name") || "").indexOf(q) !== -1;
      el.hidden = !match;
      if (match) visible += 1;
    });
    root.querySelectorAll("[data-tier]").forEach(function (tier) {
      var any = false;
      tier.querySelectorAll("[data-kind]").forEach(function (el) {
        if (!el.hidden) any = true;
      });
      tier.hidden = !any;
    });
    if (countEl) {
      if (q === "") {
        countEl.textContent = total === 1 ? "1 " + singular : total + " " + plural;
      } else {
        countEl.textContent = visible === 1 ? "1 " + singular : visible + " " + plural;
      }
    }
    if (emptyEl) emptyEl.hidden = visible !== 0 || total === 0;
  }

  function fold(s) {
    return String(s || "")
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/['’.]/g, "")
      .toLowerCase();
  }
})();
