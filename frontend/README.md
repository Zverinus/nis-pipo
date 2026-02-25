# nis-pipo frontend

Минимальный UI на React + Mantine.

## Запуск

```bash
npm install
npm run dev
```

Фронт на http://localhost:5173. **Бэк должен быть запущен** на http://localhost:8080 — Vite проксирует `/api` на него, CORS не нужен.

Для `npm run preview` или деплоя задай `VITE_API_URL` в `.env` (например `http://localhost:8080`).
