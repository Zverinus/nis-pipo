import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { TextInput, Textarea, Button, Paper, Title, Stack, Select } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { meetings } from '../api';

export default function MeetingNew() {
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dateStart, setDateStart] = useState<Date | null>(null);
  const [dateEnd, setDateEnd] = useState<Date | null>(null);
  const [slotMinutes, setSlotMinutes] = useState<string | null>('30');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!dateStart || !dateEnd || !slotMinutes) return;
    setError('');
    setLoading(true);
    try {
      const m = await meetings.create({
        title,
        description,
        date_start: dateStart!.toISOString().slice(0, 10),
        date_end: dateEnd!.toISOString().slice(0, 10),
        slot_minutes: parseInt(slotMinutes, 10),
      });
      navigate(`/meetings/${m.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper p="xl" maw={500} mx="auto" mt="xl">
      <Title order={2} mb="md">Создать встречу</Title>
      <form onSubmit={handleSubmit}>
        <Stack>
          <TextInput label="Название" value={title} onChange={(e) => setTitle(e.target.value)} required />
          <Textarea label="Описание" value={description} onChange={(e) => setDescription(e.target.value)} />
          <DatePickerInput label="Дата начала" value={dateStart} onChange={setDateStart} required />
          <DatePickerInput label="Дата окончания" value={dateEnd} onChange={setDateEnd} required />
          <Select
            label="Шаг слота (мин)"
            value={slotMinutes}
            onChange={setSlotMinutes}
            data={[
              { value: '15', label: '15 минут' },
              { value: '30', label: '30 минут' },
              { value: '60', label: '60 минут' },
            ]}
          />
          {error && <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>}
          <Button type="submit" loading={loading}>Создать</Button>
        </Stack>
      </form>
    </Paper>
  );
}
