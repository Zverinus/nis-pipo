import { useEffect, useState } from 'react';
import { Outlet, Link, useNavigate } from 'react-router-dom';
import { AppShell, Group, Title, Button, Text } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import { auth, decodeTokenPayload } from './api';

interface LayoutProps {
  token: string | null;
}

export default function Layout({ token }: LayoutProps) {
  const navigate = useNavigate();
  const [, setToken] = useLocalStorage<string | null>({ key: 'token', defaultValue: null });
  const [email, setEmail] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setEmail(null);
      return;
    }
    const payload = decodeTokenPayload(token);
    if (payload?.email) setEmail(payload.email);
    auth.me()
      .then((u: { id: string; email: string }) => setEmail(u.email))
      .catch((err) => {
        if (err instanceof Error && (err.message === 'unauthorized' || err.message.includes('401'))) {
          setToken(null);
          localStorage.removeItem('token');
          navigate('/login');
        }
      });
  }, [token, setToken, navigate]);

  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Link to="/" style={{ textDecoration: 'none', color: 'inherit' }}>
            <Title order={3}>nis-pipo</Title>
          </Link>
          {token ? (
            <Group>
              {email && (
                <Text size="sm" c="dimmed">
                  {email}
                </Text>
              )}
              <Button variant="subtle" component={Link} to="/meetings">
                Мои встречи
              </Button>
              <Button variant="subtle" component={Link} to="/meetings/new">
                Создать
              </Button>
              <Button variant="light" color="red" onClick={() => setToken(null)}>
                Выйти
              </Button>
            </Group>
          ) : (
            <Group>
              <Button variant="subtle" component={Link} to="/login">
                Войти
              </Button>
              <Button component={Link} to="/register">
                Регистрация
              </Button>
            </Group>
          )}
        </Group>
      </AppShell.Header>
      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  );
}
