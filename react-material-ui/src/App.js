import React, { Component } from "react";
import CssBaseline from "@material-ui/core/CssBaseline";

import AppBar from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import IconButton from "@material-ui/core/IconButton";
import Typography from "@material-ui/core/Typography";
import MenuIcon from "@material-ui/icons/Menu";

import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TableSortLabel from "@material-ui/core/TableSortLabel";
import Paper from "@material-ui/core/Paper";

class Header extends Component {
  render() {
    return (
      <header className="App-header">
        <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width" />
        <CssBaseline />
      </header>
    );
  }
}

class HeaderBar extends Component {
  render() {
    return (
      <div>
        <AppBar position="static">
          <Toolbar>
            <IconButton edge="start" color="inherit" aria-label="open drawer">
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" noWrap>
              Bugzilla Report
            </Typography>
          </Toolbar>
        </AppBar>
      </div>
    );
  }
}

function descendingComparator(a, b, orderBy) {
  if (b[orderBy] < a[orderBy]) {
    return -1;
  }
  if (b[orderBy] > a[orderBy]) {
    return 1;
  }
  return 0;
}

function getComparator(order, orderBy) {
  return order === "desc"
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy);
}

function stableSort(array, comparator) {
  const stabilizedThis = array.map((el, index) => [el, index]);
  stabilizedThis.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) return order;
    return a[1] - b[1];
  });
  return stabilizedThis.map(el => el[0]);
}

const BugSubComponent = props => {
  const { bug } = props;
  if (bug.sub_components) {
    return <TableCell align="center">{bug.sub_components[bug.component]}</TableCell>;
  }
  return <TableCell align="center"></TableCell>;
};

const BugTableBody = props => {
  var bugRows = [];
  Object.keys(props.bugs).map(function(key) {
    var team = props.bugs[key];
    team.map(bug => {
      bugRows.push(
        <TableRow key={bug.id}>
          <TableCell component="th" scope="row">
            {bug.id}
          </TableCell>
          <TableCell align="left">{bug.summary}</TableCell>
          <TableCell align="center">{bug.component}</TableCell>
          <BugSubComponent bug={bug} />
          <TableCell align="center">{bug.severity}</TableCell>
          <TableCell align="center">{bug.status}</TableCell>
        </TableRow>
      );
    });
  });
  return <TableBody>{bugRows}</TableBody>;
};

class BugTable extends Component {
  constructor(props) {
    super(props);
    this.state = {
      order: "asc",
      orderBy: "number"
    };
  }
  render() {
    const createSortHandler = property => event => {
      handleRequestSort(event, property);
    };
    const handleRequestSort = (event, property) => {
      const isAsc = this.state.orderBy === property && this.state.order === "asc";
      this.setState({ order: isAsc ? "desc" : "asc" });
      this.setState({ orderBy: property });
    };
    var { order, orderBy } = this.props;
    return (
      <TableContainer component={Paper}>
        <Table className="BugTable" size="small">
          <TableHead>
            <TableRow>
              <TableCell key="number" sortDirection={orderBy === "number" ? order : false}>
                <TableSortLabel
                  active={orderBy === "number"}
                  direction={orderBy === "number" ? order : "asc"}
                  onClick={createSortHandler("number")}
                >
                  Number
                  {orderBy === "number" ? (
                    <span>{order === "desc" ? "sorted descending" : "sorted ascending"}</span>
                  ) : null}
                </TableSortLabel>
              </TableCell>
              <TableCell key="summary" align="left">
                Summary
              </TableCell>
              <TableCell key="component" align="center">
                Component
              </TableCell>
              <TableCell key="subComponent" align="center">
                Sub Component
              </TableCell>
              <TableCell key="severity" align="center">
                Severity
              </TableCell>
              <TableCell key="status" align="center">
                Status
              </TableCell>
            </TableRow>
          </TableHead>
          <BugTableBody bugs={this.props.bugs} />
        </Table>
      </TableContainer>
    );
  }
}

class BZApp extends React.Component {
  state = {
    bugs: []
  };

  componentDidMount() {
    const url = "http://localhost:8000/api";
    fetch(url)
      .then(result => result.json())
      .then(result => {
        this.setState({
          bugs: result
        });
      });
  }

  removeBug = index => {
    const { bugs } = this.state;

    this.setState({
      bugs: bugs.filter((bug, i) => {
        return i !== index;
      })
    });
  };

  render() {
    const { bugs } = this.state;
    return (
      <div className="App">
        <Header />
        <HeaderBar />
        <BugTable bugs={bugs} removeBug={this.removeBug} />
      </div>
    );
  }
}

export default BZApp;
