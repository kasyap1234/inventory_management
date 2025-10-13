#!/bin/bash

# Quick Fix: Manually verify your existing account
# This allows you to login immediately without checking email

echo "🔧 Activating your account: kasyap3103@gmail.com"
echo ""

psql "postgresql://testuser:testpass@localhost:5440/testdb" -c \
  "UPDATE users SET status = 'active' WHERE email = 'kasyap3103@gmail.com' RETURNING email, status;"

if [ $? -eq 0 ]; then
  echo ""
  echo "✅ SUCCESS! Your account is now active."
  echo ""
  echo "📱 Next steps:"
  echo "   1. Go to http://localhost:3000/login"
  echo "   2. Login with:"
  echo "      Email: kasyap3103@gmail.com"
  echo "      Password: [your password]"
  echo ""
  echo "🎉 You should now be able to access the dashboard!"
else
  echo ""
  echo "❌ Failed to activate account. Please check if:"
  echo "   - PostgreSQL is running on port 5440"
  echo "   - The database credentials are correct"
  echo "   - The user exists in the database"
fi
