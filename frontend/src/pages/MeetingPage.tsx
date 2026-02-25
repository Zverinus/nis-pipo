import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Title, Paper, Stack, Table, Button, Loader, CopyButton } from '@mantine/core';
import { meetings } from '../api';
import type { Meeting, SlotResult } from '../api';

interface MeetingPageProps {
  token: string | null;
}

function slotLabel(meeting: Meeting, slotIndex: number): string {
  const start = new Date(meeting.date_start);
  const end = new Date(meeting.date_end);
  const dayCount = Math.floor((end.getTime() - start.getTime()) / (24 * 60 * 60 * 1000)) + 1;
  const slotsPerDay = (24 * 60) / meeting.slot_minutes;
  const maxSlots = dayCount * slotsPerDay;
  if (slotIndex < 0 || slotIndex >= maxSlots) return `#${slotIndex}`;

  const dayIndex = Math.floor(slotIndex / slotsPerDay);
  const timeIndex = slotIndex % slotsPerDay;
  const date = new Date(start);
  date.setDate(date.getDate() + dayIndex);
  const minutes = timeIndex * meeting.slot_minutes;
  date.setHours(Math.floor(minutes / 60), minutes % 60, 0, 0);

  const dateText = date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
  const timeText = date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  return `${dateText} ${timeText}`;
}

export default function MeetingPage({ token }: MeetingPageProps) {
  const { id } = useParams<{ id: string }>();
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [results, setResults] = useState<SlotResult[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [finalizing, setFinalizing] = useState(false);
  const [selectedSlot, setSelectedSlot] = useState<number | null>(null);

  const participantUrl = `${window.location.origin}/meetings/${id}/participate`;

  useEffect(() => {
    if (!id) return;
    meetings.get(id)
      .then(setMeeting)
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка'))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    if (!id || !token) return;
    meetings.results(id)
      .then(setResults)
      .catch(() => setResults([]));
  }, [id, token]);

  const handleFinalize = async () => {
    if (!id || selectedSlot === null) return;
    setFinalizing(true);
    try {
      await meetings.finalize(id, selectedSlot);
      setMeeting((m) => m ? { ...m, status: 'finalized', final_slot_index: selectedSlot } : null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка');
    } finally {
      setFinalizing(false);
    }
  };

  if (loading || !meeting) return <Loader m="xl" />;
  if (error) return <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>;

  const showResults = token && results !== null;
  const canFinalize = meeting.status === 'active' && showResults && selectedSlot !== null;
  const visibleResults =
    meeting.status === 'finalized' && meeting.final_slot_index !== undefined
      ? (results ?? []).filter((r) => r.slot_index === meeting.final_slot_index)
      : (results ?? []);

  return (
    <Stack>
      <Title order={2}>{meeting.title}</Title>
      <p>{meeting.description}</p>
      <p>Даты: {meeting.date_start} — {meeting.date_end}, шаг {meeting.slot_minutes} мин. Статус: {meeting.status}</p>

      {showResults && (
        <Paper p="md">
          <Title order={4} mb="md">
            {meeting.status === 'finalized'
              ? 'Итог встречи'
              : 'Результаты (кол-во участников по слотам)'}
          </Title>
          {results === null ? (
            <Loader size="sm" />
          ) : visibleResults.length === 0 ? (
            <p>Пока нет ответов</p>
          ) : (
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Время</Table.Th>
                  <Table.Th>Участников</Table.Th>
                  <Table.Th>Имена</Table.Th>
                  {meeting.status === 'active' && <Table.Th></Table.Th>}
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {visibleResults.map((r) => (
                  <Table.Tr key={r.slot_index}>
                    <Table.Td>{slotLabel(meeting, r.slot_index)}</Table.Td>
                    <Table.Td>{r.count}</Table.Td>
                    <Table.Td>{(r.participant_names ?? []).join(', ') || '—'}</Table.Td>
                    {meeting.status === 'active' && (
                      <Table.Td>
                        <Button
                          size="xs"
                          variant={selectedSlot === r.slot_index ? 'filled' : 'light'}
                          onClick={() => setSelectedSlot(r.slot_index)}
                        >
                          Выбрать
                        </Button>
                      </Table.Td>
                    )}
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          )}
          {meeting.status === 'active' && canFinalize && (
            <Button mt="md" onClick={handleFinalize} loading={finalizing}>
              Зафиксировать слот {selectedSlot}
            </Button>
          )}
        </Paper>
      )}

      <Paper p="md">
        <Title order={4} mb="sm">Ссылка для участников</Title>
        <p style={{ fontSize: 12, color: 'var(--mantine-color-dimmed)' }}>
          Отправьте эту ссылку участникам. Они введут имя и отмечат удобные слоты.
        </p>
        <CopyButton value={participantUrl}>
          {({ copied, copy }) => (
            <Button variant="light" onClick={copy}>{copied ? 'Скопировано' : 'Копировать ссылку'}</Button>
          )}
        </CopyButton>
      </Paper>
    </Stack>
  );
}
