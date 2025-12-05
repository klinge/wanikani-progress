// utils/date.js
export const yesterdayISO = () => {
  const date = new Date();
  date.setDate(date.getDate() - 1);
  return date.toISOString().split('T')[0]; // e.g. "2025-12-03"
};

export const todayISO = () => {
  const date = new Date();
  return date.toISOString().split('T')[0]; // e.g. "2025-12-03"
};