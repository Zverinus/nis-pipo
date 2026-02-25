import '@mantine/core/styles.css';
import '@mantine/dates/styles.css';
import { MantineProvider } from '@mantine/core';
import { DatesProvider } from '@mantine/dates';
import { BrowserRouter } from 'react-router-dom';
import { createRoot } from 'react-dom/client';
import App from './App';
import 'dayjs/locale/ru';

createRoot(document.getElementById('root')!).render(
  <MantineProvider defaultColorScheme="light">
    <DatesProvider settings={{ locale: 'ru' }}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </DatesProvider>
  </MantineProvider>
);
