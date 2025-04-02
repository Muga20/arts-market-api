CREATE TABLE user_details (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    gender VARCHAR(50),
    date_of_birth DATE,
    profile_image TEXT NOT NULL,
    cover_image TEXT NOT NULL,
    about_the_user TEXT,
    is_profile_public BOOLEAN NOT NULL DEFAULT false,
    nickname VARCHAR(100),
    preferred_pronouns VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_user_details_user_id ON user_details(user_id);