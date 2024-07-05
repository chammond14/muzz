package db

const seed1 = ` INSERT INTO profiles (age, name, gender, email, password, lat, long) 
				VALUES (30, 'Bob', 'male', 'bob@muzz.com', 'password', -0.13807155434153104, 51.50649673895887) ON CONFLICT DO NOTHING;`

const seed2 = ` INSERT INTO profiles (age, name, gender, email, password, lat, long) 
				VALUES (65, 'Alice', 'female', 'alice@muzz.com', 'dolphins', -0.10479690100321429, 51.50816434784823) ON CONFLICT DO NOTHING;`

const seed3 = ` INSERT INTO profiles (age, name, gender, email, password, lat, long) 
				VALUES (82, 'John', 'other', 'john@muzz.com', 'papayas', -0.13623616213333362, 38.53691669075023) ON CONFLICT DO NOTHING;`

const seed4 = ` INSERT INTO profiles (age, name, gender, email, password, lat, long) 
				VALUES (43, 'Bernadette', 'female', 'bernadette@muzz.com', 'noeledmonds', 1.6434483466408198, 52.7613366197184) ON CONFLICT DO NOTHING;`
