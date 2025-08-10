import React from 'react'
import Menu from './Menu'
import LeftSide from './LeftSide'
import RightSide from './RightSide'
import {Button, Alert, Row, Col} from 'react-bootstrap';

import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

const Home = () => {
  return (
    <div className='home'>
        <Menu/>
        
        <Row className="landing">
          <Col><LeftSide/></Col>
          <Col><RightSide/></Col>
        </Row>
      
    </div>
  )
}

export default Home
