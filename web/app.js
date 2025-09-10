async function fetchOrder() {
  const id = document.getElementById('orderId').value.trim();
  const out = document.getElementById('out');
  if (!id) { out.textContent = 'Введите order_uid'; return; }
  out.textContent = 'Загрузка…';
  try {
    const res = await fetch(`/api/order/${encodeURIComponent(id)}`);
    if (!res.ok) { out.textContent = `Ошибка: ${res.status}`; return; }
    const data = await res.json();
    out.textContent = JSON.stringify(data, null, 2);
  } catch (e) {
    out.textContent = 'Сеть недоступна или сервер не запущен';
  }
}
