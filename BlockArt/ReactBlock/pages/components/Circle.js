import React, { Component } from 'react';
import PropTypes from 'prop-types';

class Circle extends Component {
  static propTypes = {
  };

  render() {
    return (
      <svg>
      <circle
        cx={this.props.cx}
        cy={this.props.cy}
        r={this.props.r}
        stroke={this.props.stroke}
        fill={this.props.fill}
      />
      </svg>
    )
  }
}

export default Circle;
