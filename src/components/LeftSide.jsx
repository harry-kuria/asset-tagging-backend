import React, { useState } from 'react';
import {Form, Button} from 'react-bootstrap';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

const LeftSide = () => {

    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
    
        try {
          // Make a request to your backend to check the credentials
          const response = await axios.post('http://localhost:5000/api/login', {
            username,
            password,
          });
    
          // Assuming your backend responds with a success message
          if (response.data.success) {
            // alert('Login successful!');
            console.log("success");
            
            navigate('/dashboard');
          } else {
            alert('Invalid credentials. Please try again.');
          }
        } catch (error) {
          console.error('Error during login:', error);
        }
      };
 return (
     <div>
         <br/>
         <br/>
         <br/>
         <Form style={{width:"80%", marginLeft:"10%", marginTop:"10%"}}>
             <Form.Group >
                 <Form.Label>Enter your username</Form.Label>
                 <Form.Control type="text" placeholder="Enter your username" value={username}
            onChange={(e) => setUsername(e.target.value)} />
             </Form.Group>
             <Form.Group >
                 <Form.Label>Enter your password</Form.Label>
                 <Form.Control
            type="password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
             </Form.Group>
             <Button type="button" onClick={handleSubmit}>Submit</Button>
         </Form>
     </div>
 )
}

export default LeftSide;