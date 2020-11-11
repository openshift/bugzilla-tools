import React from "react";
import Container from "@material-ui/core/Container";
import Typography from "@material-ui/core/Typography";
import Box from "@material-ui/core/Box";
import ProTip from "./ProTip";
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import { Theme, createStyles, makeStyles } from "@material-ui/core/styles";
import { grey, red, green } from "@material-ui/core/colors";
import Accordion from "@material-ui/core/Accordion";
import AccordionSummary from "@material-ui/core/AccordionSummary";
import AccordionDetails from "@material-ui/core/AccordionDetails";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import WarningIcon from "@material-ui/icons/Warning";
import ThumbUpIcon from "@material-ui/icons/ThumbUp";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: "flex",
      flexWrap: "wrap",
      "& > *": {
        margin: theme.spacing(1),
        width: theme.spacing(26)
      }
    }
  })
);

// structs
// struct2ts:github.com/openshift/bugzilla-tools/pkg/slo/api.Result
interface SLOResult {
  name: string;
  current: number;
  obligation: number;
  perMember: boolean;
}

// struct2ts:github.com/openshift/bugzilla-tools/pkg/slo/api.TeamResult
interface TeamResultInterface {
  name: string;
  failing: boolean;
  members: number;
  results: SLOResult[] | null;
}

function ResultText(props: any) {
  let result = props.result;
  if (result.perMember === undefined) {
    result.perMember = false;
  }
  return (
    <Typography>
      Result: {result.current}
      <br />
      Obligation: {result.obligation}
      <br />
      PerMember: {String(result.perMember)}
    </Typography>
  );
}

function Result(props: any) {
  let result = props.result;
  let icon = <ThumbUpIcon fontSize="small" style={{ color: green[500] }} />;
  if (result.current > result.obligation) {
    icon = <WarningIcon fontSize="small" style={{ color: red[500] }} />;
  }
  return (
    <Accordion>
      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
        <Box paddingRight={1}>{icon}</Box>
        <Typography>{result.name}</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <ResultText result={result} />
      </AccordionDetails>
    </Accordion>
  );
}

function Results(props: any) {
  let results = props.results;
  if (!results) {
    return null;
  }
  let resultArray = Object.entries(results).map(([key, result]) => {
    return <Result result={result} />;
  });
  return <div>{resultArray}</div>;
}

function TeamResult(props: any) {
  let result = props.result;
  if (!result?.name) {
    return null;
  }
  let icon = <ThumbUpIcon style={{ color: green[500] }} />;
  if (result.failing) {
    icon = <WarningIcon style={{ color: red[500] }} />;
  }
  let subhdr = "Members: " + result.members;
  return (
    <Card key={result.name}>
      <CardHeader avatar={icon} title={result.name} subheader={subhdr} />
      <CardContent>
        <Results results={result.results} />
      </CardContent>
    </Card>
  );
}

interface AllTeamsProps {
  classes: any;
}

interface AllTeamsState {
  error?: any;
  isLoaded?: boolean;
  items: { [key: string]: TeamResultInterface };
  classes: any;
}

class AllTeams extends React.Component<AllTeamsProps, AllTeamsState> {
  constructor(props: AllTeamsProps) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      items: {},
      classes: props.classes
    };
  }

  componentDidMount() {
    fetch("https://team-slo-results-ocp-eng-architects.apps.ocp4.prod.psi.redhat.com/teams")
      .then(res => res.json())
      .then(
        result => {
          this.setState({
            isLoaded: true,
            items: result
          });
        },
        error => {
          this.setState({
            isLoaded: true,
            error
          });
        }
      );
  }

  render() {
    const { error, isLoaded, items, classes } = this.state;
    if (error) {
      return <div>Error: {error.message}</div>;
    } else if (!isLoaded) {
      return <div>Loading...</div>;
    } else {
      return Object.entries(items).map(([key, value]) => {
        return <TeamResult key={key} result={value} classes={classes} />;
      });
    }
  }
}

export default function App() {
  const classes = useStyles();
  return (
    <Container maxWidth="lg" style={{ backgroundColor: grey[200] }}>
      <Box my={2}>
        <Typography variant="h4" component="h1" gutterBottom align="center">
          Team SLO Results
        </Typography>
        <div className={classes.root}>
          <AllTeams classes={classes} />
        </div>
        <ProTip />
      </Box>
    </Container>
  );
}
