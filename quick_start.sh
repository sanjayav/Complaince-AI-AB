#!/bin/bash

# Quick Start Script for LangChain RAG Service
# This script gets you up and running in minutes

set -e

echo "🚀 Quick Start for LangChain RAG Service"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Python 3 is not installed${NC}"
    echo "Please install Python 3.9+ and try again"
    exit 1
fi

# Check if pip is available
if ! command -v pip3 &> /dev/null; then
    echo -e "${RED}❌ pip3 is not installed${NC}"
    echo "Please install pip3 and try again"
    exit 1
fi

echo -e "${GREEN}✅ Python and pip found${NC}"

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo -e "${YELLOW}📦 Creating virtual environment...${NC}"
    python3 -m venv venv
    echo -e "${GREEN}✅ Virtual environment created${NC}"
else
    echo -e "${GREEN}✅ Virtual environment already exists${NC}"
fi

# Activate virtual environment
echo -e "${YELLOW}🔧 Activating virtual environment...${NC}"
source venv/bin/activate
echo -e "${GREEN}✅ Virtual environment activated${NC}"

# Install dependencies
echo -e "${YELLOW}📚 Installing dependencies...${NC}"
pip install -r requirements.txt
echo -e "${GREEN}✅ Dependencies installed${NC}"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}⚙️  Creating .env file from template...${NC}"
    if [ -f "env.example" ]; then
        cp env.example .env
        echo -e "${GREEN}✅ .env file created from env.example${NC}"
        echo -e "${YELLOW}⚠️  Please edit .env file with your actual credentials${NC}"
        echo -e "${YELLOW}   Required: OPENAI_API_KEY, QDRANT_URL, QDRANT_API_KEY${NC}"
    else
        echo -e "${RED}❌ env.example file not found${NC}"
        echo "Please create a .env file manually with your credentials"
    fi
else
    echo -e "${GREEN}✅ .env file already exists${NC}"
fi

# Run setup verification
echo -e "${YELLOW}🔍 Running setup verification...${NC}"
python setup_langchain_rag.py

echo ""
echo -e "${GREEN}🎉 Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your credentials"
echo "2. Run: python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8000"
echo "3. Open: http://localhost:8000/docs"
echo ""
echo "Need help? Run: python setup_langchain_rag.py"
