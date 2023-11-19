import React, { FormEvent, ReactElement, useEffect } from 'react';
import { Helmet } from 'react-helmet-async';
import { useLocation } from 'react-router';
import { useNavigate } from 'react-router-dom';
import Path from '@domain/constants/Path';
import useAuth from '@hooks/useAuth';

export type LoginFormValues = {
  email_address: string;
  password: string;
};

type LocationState = { from?: Location };

const LoginPage = (): ReactElement => {
  const authService = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (authService.isLoggedIn()) {
      navigate(Path.TEMP, { replace: true });
    }
  }, [authService, navigate]);

  const location = useLocation();
  const state: LocationState = location.state as LocationState;

  const handleLogin = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const emailAddress = formData.get('email_address')?.toString() ?? '';
    const password = formData.get('password')?.toString() ?? '';
    await authService.login({ email_address: emailAddress, password });
    const from = state?.from?.pathname || Path.TEMP;
    navigate(from, { replace: true });
  };

  return (
    <div className="login-page">
      <Helmet>
        <title>Pascal Allen - Login</title>
      </Helmet>
      <header className="header">
        <h1>Login</h1>
      </header>
      <section className="login-form-section">
        <form className="login-form" onSubmit={handleLogin}>
          <div className="form-group">
            <label htmlFor="email-address">Email address</label>
            <input id="email-address" className="email-address" type="email" name="email_address" />
          </div>
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input id="password" className="password" type="password" name="password" />
          </div>
          <div className="form-group">
            <button id="submit" className="submit" type="submit">
              Submit
            </button>
          </div>
        </form>
      </section>
    </div>
  );
};

export default LoginPage;
