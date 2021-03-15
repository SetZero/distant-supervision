import React from 'react';
import logo from './logo.svg';
import './App.css';
import { CallComponent } from './components/CallComponent';
import { createMuiTheme, makeStyles, ThemeProvider } from '@material-ui/core/styles';
import { useDispatch, useSelector } from "react-redux";
import { useEffect } from 'react';

function App() {
  const prefersDarkMode = useSelector((state: any) => state.darkMode);

  const theme = React.useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: prefersDarkMode ? 'dark' : 'light',
        },
      }),
    [prefersDarkMode],
  );

  const useStyles = makeStyles({
    root: {
      background: theme.palette.background.default,
      height: '100%'
    }
  });
  const classes = useStyles();

  return (
    <ThemeProvider theme={theme}>
      <div className="App">
        <section className={classes.root}>
          <CallComponent />
        </section>
      </div>
    </ThemeProvider>
  );
}

export default App;
