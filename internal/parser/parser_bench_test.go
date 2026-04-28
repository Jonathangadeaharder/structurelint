package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParseFile_Go(b *testing.B) {
	goFile := filepath.Join("..", "..", "cmd", "structurelint", "main.go")
	if _, err := os.Stat(goFile); err != nil {
		b.Skipf("main.go not found: %v", err)
	}

	parser := New(filepath.Join("..", ".."))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(goFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

func BenchmarkParseFile_TypeScript(b *testing.B) {
	tmpDir := b.TempDir()
	tsFile := filepath.Join(tmpDir, "component.ts")

	content := `import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import axios from 'axios';
import { ThemeProvider, createTheme } from '@mui/material';
import { makeStyles } from '@mui/styles';
import { format, parseISO } from 'date-fns';
import { debounce, throttle } from 'lodash-es';
import { v4 as uuidv4 } from 'uuid';
import useSWR from 'swr';
import { QueryClient, useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as Yup from 'yup';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { useMediaQuery } from '@mui/material';
import styled from '@emotion/styled';
import { keyframes } from '@emotion/react';
import { CSSTransition, TransitionGroup } from 'react-transition-group';
import { FixedSizeList as List } from 'react-window';
import InfiniteScroll from 'react-infinite-scroll-component';
import Dropzone from 'react-dropzone';
import { Editor } from '@monaco-editor/react';
import { DiffEditor } from '@monaco-editor/react';
import Prism from 'prismjs';
import 'prismjs/components/prism-typescript';
import 'prismjs/themes/prism-tomorrow.css';
import { Toaster, toast } from 'sonner';
import { Dialog, Transition } from '@headlessui/react';
import { Fragment } from 'react';
import { Menu, Transition as MenuTransition } from '@headlessui/react';
import { SearchIcon, ChevronDownIcon } from '@heroicons/react/outline';
import { PlusIcon, TrashIcon, PencilIcon } from '@heroicons/react/solid';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { ExclamationTriangleIcon } from '@heroicons/react/24/solid';
import './styles.css';
import './animations.scss';
import '../utils/helpers';
import '../../types/global';
import './components/Button';
import './hooks/useDebounce';
import './context/AuthContext';
import './services/api';
import './utils/formatDate';
import './lib/constants';
import './store/useStore';
import './middleware/auth';
import './decorators/withAuth';
import './hocs/withTheme';
import './mixins/responsive';
import './modules/feature-flags';
import './plugins/analytics';
import './providers/QueryProvider';
import './routes/index';
import './schemas/validation';
import './validators/email';
import './workers/processing';
`

	if err := os.WriteFile(tsFile, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	parser := New(tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(tsFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

func BenchmarkParseFile_Python(b *testing.B) {
	tmpDir := b.TempDir()
	pyFile := filepath.Join(tmpDir, "module.py")

	content := `import os
import sys
import json
import hashlib
import asyncio
import logging
from typing import List, Dict, Optional, Tuple, Any
from datetime import datetime, timedelta
from pathlib import Path
from collections import defaultdict, OrderedDict
from functools import lru_cache, wraps
from dataclasses import dataclass, field
from abc import ABC, abstractmethod
import numpy as np
import pandas as pd
from sqlalchemy import Column, Integer, String, ForeignKey, create_engine
from sqlalchemy.orm import sessionmaker, relationship, declarative_base
from flask import Flask, request, jsonify, g
from fastapi import FastAPI, Depends, HTTPException, status
from pydantic import BaseModel, Field, validator
import redis
import celery
from celery import Celery
import requests
from bs4 import BeautifulSoup
from scrapy import Spider, Item, Field as ScrapyField
import boto3
from google.cloud import storage, bigquery
from prometheus_client import Counter, Histogram, Gauge
import structlog
from .models import User, Product, Order
from .services.auth import authenticate, authorize
from .services.payment import process_payment
from .utils import validate_email, format_currency
from .config import settings
from .tasks import send_email, process_data
`

	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	parser := New(tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(pyFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

func BenchmarkParseExports_Go(b *testing.B) {
	goFile := filepath.Join("..", "..", "cmd", "structurelint", "main.go")
	if _, err := os.Stat(goFile); err != nil {
		b.Skipf("main.go not found: %v", err)
	}

	parser := New(filepath.Join("..", ".."))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseExports(goFile)
		if err != nil {
			b.Fatalf("ParseExports failed: %v", err)
		}
	}
}

func BenchmarkParseFile_Java(b *testing.B) {
	tmpDir := b.TempDir()
	javaFile := filepath.Join(tmpDir, "Service.java")

	content := `package com.example.application.service;

import com.example.domain.model.User;
import com.example.domain.model.Product;
import com.example.domain.repository.UserRepository;
import com.example.domain.repository.ProductRepository;
import com.example.infrastructure.persistence.DatabaseConnection;
import com.example.presentation.dto.UserResponse;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.cache.annotation.CacheEvict;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import java.util.UUID;
import java.time.LocalDateTime;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import javax.validation.Valid;
import javax.persistence.EntityManager;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserService {
    private final UserRepository userRepository;
    private final ProductRepository productRepository;
    private final DatabaseConnection dbConnection;
    private final EntityManager entityManager;
}
`

	if err := os.WriteFile(javaFile, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	parser := New(tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(javaFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

func BenchmarkResolveImportPath(b *testing.B) {
	parser := New("/project")

	sources := []string{
		"src/app.ts",
		"src/components/Button.tsx",
		"src/deep/nested/path/module.ts",
	}
	imports := []string{
		"./utils",
		"../hooks/useButton",
		"../../lib/constants",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, src := range sources {
			for _, imp := range imports {
				parser.ResolveImportPath(src, imp)
			}
		}
	}
}
