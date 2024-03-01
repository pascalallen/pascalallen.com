import React, { FormEvent, ReactElement, useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { Helmet } from 'react-helmet-async';
import { useNavigate } from 'react-router-dom';
import Path from '@domain/constants/Path';
import useStore from '@hooks/useStore';
import AuthService from '@services/AuthService';
import TempService from '@services/TempService';

const TempPage = observer((): ReactElement => {
  const authStore = useStore('authStore');
  const navigate = useNavigate();

  const [eventStreamMessage, setEventStreamMessage] = useState();

  useEffect(() => {
    const eventSource = new EventSource('/api/v1/event-stream');
    eventSource.addEventListener('message', event => {
      setEventStreamMessage(event.data);
    });

    eventSource.onerror = () => {
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, []);

  useEffect(() => {
    const tempService = new TempService(authStore);
    tempService.temp().then(
      r => console.log(r),
      e => console.error(e)
    );
  }, [authStore]);

  const handleEventStreamPost = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const message = formData.get('message')?.toString() ?? '';
    const tempService = new TempService(authStore);
    await tempService.eventStreamPost({ message });
  };

  const handleLogout = async (): Promise<void> => {
    const authService = new AuthService(authStore);
    authService.logout().finally(() => navigate(Path.INDEX));
  };

  return (
    <div className="temp-page">
      <Helmet>
        <title>Pascal Allen - Temp</title>
      </Helmet>
      <header className="header">
        <h1>You&apos;re authenticated and connected to the event stream at /api/v1/event-stream!</h1>
        <h2>Event stream message received: {eventStreamMessage}</h2>
      </header>
      <section className="temp-form-section">
        <form className="temp-form" onSubmit={handleEventStreamPost}>
          <div className="form-group">
            <label htmlFor="message">Message</label>
            <input id="event-stream-message-input" className="event-stream-message-input" type="text" name="message" />
          </div>
          <div className="form-group">
            <button id="submit-temp-form-button" className="submit" type="submit">
              Post to event stream
            </button>
          </div>
        </form>
        <br />
        <button type="button" onClick={handleLogout}>
          Logout
        </button>
      </section>
    </div>
  );
});

export default TempPage;
