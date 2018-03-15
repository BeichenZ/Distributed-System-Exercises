import React, { Component } from 'react';
import SVG from './components/SVG'
import Circle from './components/Circle'
import Head from 'next/head';



class BlockartSVG extends Component {
  constructor(props) {
    super(props);

    this.state = {
      paths: []
    };
  }

  componentDidMount(){
    setInterval(this.periodicFetchSVG, 1000);
  }

  periodicFetchSVG = () => {
    fetch("http://localhost:5000/getshapes", {
      method: 'GET'
    })
    .then(res => res.json())
    .then(response => {
      console.log('this is the response')
      console.log(response)
      this.setState({paths: response})
    })
    .catch(error => console.error('Error:', error))
  }

  addShapes = () => {
    fetch("http://localhost:5000/addshape", {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        shape: 'circle',
      })
    })
  }

  renderSVG = () => {
    const svgs = this.state.paths.map((svg, index) => {
      if (svg.Path[0] == "c") {
        const circleArr = svg.Path.split(" ")
        return (
          <span>
            <Circle id={index} cx={circleArr[1]} cy={circleArr[3]} r={circleArr[5]} fill={svg.Fill} stroke={svg.Stroke}/>
          </span>
        )
      } else {
        return (
          <span>
            <SVG id={index} d={svg.Path} fill={svg.Fill} stroke={svg.Stroke}/>
          </span>
        )
      }
    });
    return svgs
  }


  render() {
    return (
      <div>
       <Head>
         <title>My styled page</title>
         <link href="./css/flex.css" rel="stylesheet" />
       </Head>

         <p>
           CPSC 416
         </p>
         <div>
           <button onClick={this.addShapes.bind(this)}></button>
           <div>
             {this.renderSVG()}
           </div>
         </div>
     </div>

    )
  }
}

export default BlockartSVG
