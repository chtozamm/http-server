CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    author VARCHAR(100),
    message TEXT
);

INSERT INTO posts (author, message) VALUES
('Bilbo Beggins', 'Let the adventure begin...'),
('Obi-Wan Kenobi', 'Hello there!'),
('Geralt of Rivia', 'Hmmm... Wind''s howling...'),
('Obi-Wan Kenobi', 'May the force be with you ⚡'),
('R2-D2', 'May the 4th bla-bla bee-boop');
