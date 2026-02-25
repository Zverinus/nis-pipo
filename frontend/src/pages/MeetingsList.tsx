import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Table, Button, Title, Stack, Loader, Group } from '@mantine/core';
import { meetings } from '../api';
import type { Meeting } from '../api';

export default function MeetingsList() {
  const [list, setList] = useState<Meeting[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    meetings.list()
      .then(setList)
      .catch((err) => setError(err instanceof Error ? err.message : 'Ошибка'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Loader m="xl" />;
  if (error) return <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>;

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Мои встречи</Title>
        <Button component={Link} to="/meetings/new">Создать встречу</Button>
      </Group>
      {list.length === 0 ? (
        <p>Нет встреч. <Link to="/meetings/new">Создать первую</Link></p>
      ) : (
        <Table>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Название</Table.Th>
              <Table.Th>Даты</Table.Th>
              <Table.Th>Статус</Table.Th>
              <Table.Th></Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {list.map((m) => (
              <Table.Tr key={m.id}>
                <Table.Td>{m.title}</Table.Td>
                <Table.Td>{m.date_start} — {m.date_end}</Table.Td>
                <Table.Td>{m.status}</Table.Td>
                <Table.Td>
                  <Button variant="subtle" component={Link} to={`/meetings/${m.id}`}>
                    Открыть
                  </Button>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      )}
    </Stack>
  );
}

