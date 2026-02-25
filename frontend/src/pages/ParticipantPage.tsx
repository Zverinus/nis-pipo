import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Title, Paper, Stack, TextInput, Button, Loader, Table, Text } from '@mantine/core';
import { meetings, participants } from '../api';
import type { Meeting } from '../api';

const WEEKDAY = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

function parseDateStart(s: string): Date {
  if (s.includes('T')) return new Date(s);
  const [y, m, d] = s.split('-').map(Number);
  return new Date(y, m - 1, d);
}

/** Слот-индекс → дата и время (для подписей). */
function slotToDateTime(m: Meeting, slotIndex: number): { date: Date; timeLabel: string } {
  const slotsPerDay = (24 * 60) / m.slot_minutes;
  const dayIndex = Math.floor(slotIndex / slotsPerDay);
  const timeIndex = slotIndex % slotsPerDay;
  const start = parseDateStart(m.date_start);
  const date = new Date(start);
  date.setDate(date.getDate() + dayIndex);
  const minutes = timeIndex * m.slot_minutes;
  date.setHours(Math.floor(minutes / 60), minutes % 60, 0, 0);
  const timeLabel = date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  return { date, timeLabel };
}

function slotToDayLabel(m: Meeting, slotIndex: number): string {
  const slotsPerDay = (24 * 60) / m.slot_minutes;
  const dayIndex = Math.floor(slotIndex / slotsPerDay);
  const start = parseDateStart(m.date_start);
  const date = new Date(start);
  date.setDate(date.getDate() + dayIndex);
  return `${WEEKDAY[date.getDay()]} ${date.getDate()}`;
}

export default function ParticipantPage() {
  const { id, token: tokenParam } = useParams<{ id: string; token?: string }>();
  const navigate = useNavigate();
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [displayName, setDisplayName] = useState('');
  const [participantToken, setParticipantToken] = useState<string | null>(tokenParam ?? null);

  useEffect(() => {
    if (tokenParam) setParticipantToken(tokenParam);
  }, [tokenParam]);
  const [selectedSlots, setSelectedSlots] = useState<Set<number>>(new Set());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saveStatus, setSaveStatus] = useState<'idle' | 'dirty' | 'saving' | 'saved'>('idle');

  const toggleSlot = (idx: number) => {
    setSelectedSlots((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
    setSaveStatus('dirty');
  };

  const handleJoin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !displayName.trim()) return;
    setError('');
    setLoading(true);
    try {
      const { token: t } = await participants.create(id, displayName.trim());
      setParticipantToken(t);
      navigate(`/meetings/${id}/participate/${t}`, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSlots = async () => {
    if (!id || !participantToken) return;
    setSaving(true);
    setError('');
    setSaveStatus('saving');
    try {
      await participants.setSlots(id, participantToken, Array.from(selectedSlots).sort((a, b) => a - b));
      setSaveStatus('saved');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка');
      setSaveStatus('dirty');
    } finally {
      setSaving(false);
    }
  };

  useEffect(() => {
    if (!id) return;
    meetings.get(id)
      .then(setMeeting)
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка'))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading && !meeting) return <Loader m="xl" />;
  if (error && !meeting) return <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>;
  if (!meeting) return null;

  if (meeting.status === 'finalized') {
    return (
      <Paper p="xl" maw={400} mx="auto" mt="xl">
        <Title order={3}>Встреча завершена</Title>
        <p>Организатор уже зафиксировал итоговое время.</p>
      </Paper>
    );
  }

  if (!participantToken) {
    return (
      <Paper p="xl" maw={400} mx="auto" mt="xl">
        <Title order={2} mb="md">Участие во встрече</Title>
        <Text mb="md">{meeting.title}</Text>
        <form onSubmit={handleJoin}>
          <Stack>
            <TextInput
              label="Ваше имя"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Как к вам обращаться"
              required
            />
            {error && <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>}
            <Button type="submit" loading={loading}>Присоединиться</Button>
          </Stack>
        </form>
      </Paper>
    );
  }

  const slotsPerDay = (24 * 60) / meeting.slot_minutes;
  const start = parseDateStart(meeting.date_start);
  const end = new Date(meeting.date_end);
  const dayCount = Math.floor((end.getTime() - start.getTime()) / (24 * 60 * 60 * 1000)) + 1;

  return (
    <Stack>
      <Title order={2}>{meeting.title}</Title>
      <Text size="sm" c="dimmed">Отметьте удобные слоты (клик по ячейке). Колонки — дни, строки — время.</Text>
      <Table withTableBorder withColumnBorders layout="fixed">
        <Table.Thead>
          <Table.Tr>
            <Table.Th style={{ width: 56 }}>Время</Table.Th>
            {Array.from({ length: dayCount }, (_, dayIndex) => {
              const slotIndex = dayIndex * slotsPerDay;
              return (
                <Table.Th key={dayIndex} style={{ minWidth: 64 }}>
                  {slotToDayLabel(meeting, slotIndex)}
                </Table.Th>
              );
            })}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {Array.from({ length: slotsPerDay }, (_, timeIndex) => (
            <Table.Tr key={timeIndex}>
              <Table.Td style={{ fontWeight: 500 }}>
                {slotToDateTime(meeting, timeIndex).timeLabel}
              </Table.Td>
              {Array.from({ length: dayCount }, (_, dayIndex) => {
                const slotIndex = dayIndex * slotsPerDay + timeIndex;
                const chosen = selectedSlots.has(slotIndex);
                return (
                  <Table.Td key={dayIndex} style={{ padding: 2 }}>
                    <Button
                      fullWidth
                      variant={chosen ? 'filled' : 'light'}
                      size="xs"
                      color={chosen ? 'green' : 'gray'}
                      onClick={() => toggleSlot(slotIndex)}
                    >
                      {chosen ? '✓' : ''}
                    </Button>
                  </Table.Td>
                );
              })}
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
      {saveStatus === 'dirty' && (
        <Text size="sm" c="yellow">
          Есть несохраненные изменения
        </Text>
      )}
      {saveStatus === 'saving' && (
        <Text size="sm" c="dimmed">
          Сохраняем выбор...
        </Text>
      )}
      {saveStatus === 'saved' && (
        <Text size="sm" c="green">
          Выбор сохранен
        </Text>
      )}
      {error && <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>}
      <Button onClick={handleSaveSlots} loading={saving}>Сохранить</Button>
    </Stack>
  );
}
