import React from "react";

import Box from "@material-ui/core/Box";
import Grid from "@material-ui/core/Grid";
import Button from "@material-ui/core/Button";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Container from "@material-ui/core/Container";
import Link from "@material-ui/core/Link";
import Typography from "@material-ui/core/Typography";
import Paper from "@material-ui/core/Paper";
import { Theme, createStyles, makeStyles } from "@material-ui/core/styles";
import { grey, red, blue } from "@material-ui/core/colors";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      display: 'flex',
      //flexGrow: 1,
      flexWrap: 'wrap',
      '& > *': {
        margin: theme.spacing(1),
        backgroundColor: red[200],
      },
    },
    inside: {
      display: 'flex',
      //flexGrow: 1,
      flexWrap: 'wrap',
      '& > *': {
        //margin: theme.spacing(1),
        //width: theme.spacing(16),
        //height: theme.spacing(16),
        backgroundColor: grey[200],
      },
    },
  }),
);

function VersionMenu(props: any) {
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleClickListItem = (event: any) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuItemClick = (event: any, index: number) => {
    props.setversionindex(index);
    setAnchorEl(null);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <Paper elevation={3}>
      <List component="nav">
        <ListItem button onClick={handleClickListItem}>
          <ListItemText primary={props.name} secondary={props.versions[props.versionindex]} />
        </ListItem>
      </List>
      <Menu anchorEl={anchorEl} keepMounted open={Boolean(anchorEl)} onClose={handleClose} >
         {props.versions.map((option: any, index: number) => (
           <MenuItem key={option} selected={index === props.versionindex} onClick={(event) => handleMenuItemClick(event, index)}>
            {option}
          </MenuItem>
        ))}
      </Menu>
    </Paper>
  );
}

function DiffLink(props: any) {
  const link = "https://issues.redhat.com/browse/" + props.issuekey;
  return (
    <Typography variant={props.variant} component={props.component}>
      {props.prefix}:&nbsp;
      <Link href={link}>
        {props.linktext}
      </Link>
    </Typography>
  );
}

function DiffIssueTitle(props: any) {
  return (
    <DiffLink variant={props.variant} compoent={props.component} prefix="Title" linktext={props.summary} issuekey={props.issuekey} />
  );
}

function DiffIssueCard(props: any) {
  return (
    <DiffLink variant={props.variant} compoent={props.component} prefix="Issue" linktext={props.issuekey} issuekey={props.issuekey} />
  );
};

function DiffIssuePlanningLabels(props: any) {
  return (
    <Typography variant={props.variant} component={props.component}>
      Planning Labels: {props.planninglabels.join(",")}
    </Typography>
  );
}

function DiffIssueFixedVersions(props: any) {
  return (
    <Typography variant={props.variant} component={props.component}>
      Fixed Versions: {props.fixedversions.join(",")}
    </Typography>
  );
}

function DiffIssueStatus(props: any) {
  return (
    <Typography variant={props.variant} component={props.component}>
      Status: {props.status}
    </Typography>
  );
}

function DiffIssue(props: any) {
  return (
    <Card key={props.issue.key} variant="outlined">
      <CardContent>
        <DiffIssueTitle variant="body2" component="p" summary={props.issue.summary} issuekey={props.issue.key} />
        <DiffIssueCard variant="body2" component="p" issuekey={props.issue.key} />
        <DiffIssueStatus variant="body2" component="p" status={props.issue.status} />
        <DiffIssuePlanningLabels variant="body2" component="p" planninglabels={props.issue.planninglabels} />
        <DiffIssueFixedVersions variant="body2" component="p" fixedversions={props.issue.fixedversions} />
      </CardContent>
    </Card>
  );
}

function DiffCard(props: any) {
  let title = props.title
  let issues = props.issues
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h5" component="h2" color="textSecondary" gutterBottom>
          {title}
        </Typography>
        {issues.map((issue: any, index: number) => (
          <DiffIssue key={issue.key} issue={issue} />
        ))}
      </CardContent>
    </Card>
  );
}

function DiffDisplay(props: any) {
  if (props.diffing) {
    return (<Paper>Calculating Results....</Paper>);
  }
  if (props.error) {
    return ( <Paper>{props.error.message}</Paper>);
  }
  if (!props.diff) {
    return (<Paper>No Results To Display</Paper>);
  }
  return (
    <Grid container spacing={3}>
      <Grid item xs={6}>
        <Paper>
          <DiffCard title="Added" issues={props.diff.added} />
        </Paper>
      </Grid>
      <Grid item xs={6}>
        <Paper>
          <DiffCard title="Removed" issues={props.diff.removed} />
        </Paper>
      </Grid>
    </Grid>
  );
}

function TestVersionsMenu(props: any) {
  const [error, setError] = React.useState<any>(null);
  const [diffError, setDiffError] = React.useState<any>(null);
  const [isLoaded, setIsLoaded] = React.useState(false);
  const [snapshots, setSnapshots] = React.useState<any>(null);
  const [oldver, setOldVer] = React.useState(0);
  const [newver, setNewVer] = React.useState(0);
  const [diff, setDiff] = React.useState<any>(null);
  const [diffing, setDiffing] = React.useState(false);
  let host = "https://jira-daily-diff-ocp-eng-architects.apps.ocp4.prod.psi.redhat.com/";
  if (process.env.NODE_ENV == "development") {
    host = "http://localhost:8002/";
  }

  const handleDiffClick = () => {
    let oldDiff = snapshots.snapshots[oldver]
    let newDiff = snapshots.snapshots[newver]
    const url = host + "diff?oldDate=" + oldDiff + "&newDate=" + newDiff;
    setDiff(null);
    setDiffing(true);
    console.log(url);
    fetch(url)
      .then(res => res.json())
      .then(
        result => {
          setDiff(result);
          setDiffing(false);
          setDiffError(null);
        },
        error => {
          console.log(error);
          setDiff(null);
          setDiffing(false);
          setDiffError(error);
        }
      );
  };

  React.useEffect(() => {
    if (isLoaded) {
      return
    }
    const url = host + "snapshots"
    console.log(url);
    fetch(url)
      .then(res => res.json())
      .then(
        result => {
          setSnapshots(result);
          var snaplen = result.snapshots.length
          setNewVer(snaplen-1);
          setOldVer(snaplen-2);
          setIsLoaded(true);
        },
        error => {
          setError(error);
          setIsLoaded(true);
        }
      );
  });

  if (!isLoaded) {
    return <div>Loading...</div>;
  } else if (error) {
    console.log(error);
    return <div>Error: {error.message}</div>;
  } else {
    return (
      <Grid container className={props.classes.inside} spacing={3}>
        <Grid item xs={6}>
          <VersionMenu name="Old Version" classes={props.classes} versions={snapshots.snapshots} versionindex={oldver} setversionindex={setOldVer} />
        </Grid>
        <Grid item xs={6}>
          <VersionMenu name="New Version" classes={props.classes} versions={snapshots.snapshots} versionindex={newver} setversionindex={setNewVer} />
        </Grid>
        <Grid item xs={12}>
          <Box display="flex" justifyContent="center">
            <Button color="primary" variant="contained" onClick={() => handleDiffClick()}>Click Here To Calculate!</Button>
          </Box>
        </Grid>
        <Grid item xs={12}>
          <DiffDisplay classes={props.classes} diff={diff} diffing={diffing} error={diffError}/>
        </Grid>
      </Grid>
    );
  }
}

    //<Container style={{ backgroundColor: grey[200] }}>
    //</Container>
export default function App() {
  const classes = useStyles();
  return (
    <Box className={classes.root}>
      <Grid container className={classes.inside} spacing={3}>
        <Grid item xs={12}>
          <Typography variant="h4" component="h1" gutterBottom align="center">
            Jira State Diff
          </Typography>
        </Grid>
        <Grid item xs={12}>
        <TestVersionsMenu classes={classes} />
        </Grid>
      </Grid>
    </Box>
  );
}
        //<div className={classes.inside}>
        //</div>
