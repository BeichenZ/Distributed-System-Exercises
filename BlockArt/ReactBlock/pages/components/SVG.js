import React, { Component } from 'react';
import PropTypes from 'prop-types';

class SVG extends Component {
  static propTypes = {
  };

  render() {
    return (
      <svg>
        <path d={this.props.d} fill={this.props.fill} stroke={this.props.stroke} stroke-width="3"/>
      </svg>
    )
  }
}

export default SVG;
