import React from 'react';
import {Image} from "react-bootstrap";
import bgimage from "../img/bg.jpg";

const RightSide = () => {
    return (
        <div>
           <Image src={bgimage} thumbnail style={{border:"none"}} /> 
        </div>
    )
}

export default RightSide;