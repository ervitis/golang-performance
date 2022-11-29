package common

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type fileOperator struct {
	files     []*os.File
	paths     []string
	writeFile *os.File
}

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Gender    string
	Country   string
	Birthday  time.Time
}

type Users []User

type FileOperator interface {
	OpenFiles(folderPath string) error
	Read() (Users, error)
	Write(Users) error

	OpenFilesGoRoutine(folderPath string) <-chan error
	ReadGoRoutine() Users
	WriteGoRoutine(users Users) error
}

func NewFileOperator(writePathFile string) FileOperator {
	wt, err := os.Create(writePathFile)
	if err != nil {
		panic(fmt.Errorf("FileOperator: create temp file for writing error: %w", err))
	}

	fo := &fileOperator{writeFile: wt}
	return fo
}

func (f *fileOperator) OpenFiles(folderPath string) error {
	p, err := filepath.Abs(folderPath)
	if err != nil {
		return fmt.Errorf("FileOperator: opening files from absolute path %s: %w", folderPath, err)
	}

	// we don't need to follow symlinks
	if err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if len(f.paths) == 0 {
			f.paths = make([]string, 0)
		}
		f.paths = append(f.paths, path)
		return nil
	}); err != nil {
		return fmt.Errorf("FileOperator: walk inside directory %s: %w", folderPath, err)
	}

	if len(f.files) == 0 {
		f.files = make([]*os.File, len(f.paths), len(f.paths))
	}

	for i, p := range f.paths {
		t, err := os.Open(p)
		if err != nil {
			return err
		}

		f.files[i] = t
	}

	return nil
}

func (f *fileOperator) Read() (Users, error) {
	var users Users
	for _, fd := range f.files {
		r := csv.NewReader(fd)
		r.Comment = '#'
		r.Comma = ','

		d, err := r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("FileOperator: read file %s error: %w", fd.Name(), err)
		}

		if len(d) == 0 {
			return nil, nil
		}

		users = make(Users, len(d)-1, len(d))
		for i, l := range d {
			if i == 0 {
				continue
			}
			if len(l) == 0 {
				continue
			}

			id, _ := strconv.Atoi(l[0])
			td, _ := time.Parse("2006/01/02", l[6])
			users[i-1] = User{
				ID:        id,
				FirstName: l[1],
				LastName:  l[2],
				Email:     l[3],
				Gender:    l[4],
				Country:   l[5],
				Birthday:  td,
			}
		}
	}
	return users, nil
}

// Write everything in one file
func (f *fileOperator) Write(users Users) error {
	w := csv.NewWriter(f.writeFile)
	d := make([][]string, 0)
	for _, user := range users {
		d = append(d, user.ToArray())
	}
	if err := w.WriteAll(d); err != nil {
		return fmt.Errorf("FileOperator: write data in the file: %w", err)
	}
	return nil
}

func (f *fileOperator) WriteGoRoutine(users Users) error {
	w := csv.NewWriter(f.writeFile)

	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(users))
	for _, user := range users {
		go func(user User) {
			defer wg.Done()
			mtx.Lock()
			defer mtx.Unlock()
			if err := w.Write(user.ToArray()); err != nil {
				return
			}
		}(user)
	}
	wg.Wait()
	return nil
}

func (u User) ToArray() []string {
	return []string{strconv.Itoa(u.ID), fmt.Sprintf("%s %s", u.FirstName, u.LastName), u.Email, u.Country, u.Gender, u.Birthday.Format("2006-01-02")}
}

func (f *fileOperator) ReadGoRoutine() Users {
	if len(f.files) == 0 {
		return nil
	}

	users := make(Users, 0)

	wg := sync.WaitGroup{}
	wg.Add(len(f.files))
	mtx := sync.Mutex{}

	for _, fd := range f.files {
		go func(fd *os.File, users *Users) {
			defer wg.Done()

			r := csv.NewReader(fd)
			r.Comment = '#'
			r.Comma = ','

			d, err := r.ReadAll()
			if err != nil {
				return
			}

			for i, l := range d {
				if i == 0 {
					continue
				}
				id, _ := strconv.Atoi(l[0])
				td, _ := time.Parse("2006/01/02", l[6])
				mtx.Lock()
				*users = append(*users, User{
					ID:        id,
					FirstName: l[1],
					LastName:  l[2],
					Email:     l[3],
					Gender:    l[4],
					Country:   l[5],
					Birthday:  td,
				})
				mtx.Unlock()
			}
		}(fd, &users)
	}

	wg.Wait()
	return users
}

func (f *fileOperator) OpenFilesGoRoutine(folderPath string) <-chan error {
	errCh := make(chan error)
	defer close(errCh)

	p, err := filepath.Abs(folderPath)
	if err != nil {
		errCh <- fmt.Errorf("FileOperator: opening files from absolute path %s: %w", folderPath, err)
		return errCh
	}

	// we don't need to follow symlinks
	if err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if len(f.paths) == 0 {
			f.paths = make([]string, 0)
		}
		f.paths = append(f.paths, path)

		return nil
	}); err != nil {
		errCh <- fmt.Errorf("FileOperator: walk inside directory %s: %w", folderPath, err)
	}

	if len(f.files) == 0 {
		f.files = make([]*os.File, 0)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(f.paths))
	mtx := sync.Mutex{}

	for _, p := range f.paths {
		go func(p string) {
			defer wg.Done()
			mtx.Lock()
			defer mtx.Unlock()
			t, err := os.Open(p)
			if err != nil {
				errCh <- err
				log.Printf("Error: file %s: err:%v\n", p, err)
				return
			}
			f.files = append(f.files, t)
		}(p)
	}

	wg.Wait()

	return errCh
}

type UserStream struct {
	file  *os.File
	users Users
}

type fileOperatorChannel struct {
	writeFile *os.File
}

type FileOperatorChannel interface {
	OpenFiles(folderPath string) <-chan UserStream
	Read(in <-chan UserStream) <-chan UserStream
	Write(done chan struct{}, in <-chan UserStream)
}

func NewFileOperatorChannel(writePathFile string) FileOperatorChannel {
	wt, err := os.Create(writePathFile)
	if err != nil {
		panic(fmt.Errorf("FileOperator: create temp file for writing error: %w", err))
	}

	fo := &fileOperatorChannel{writeFile: wt}
	return fo
}

func (f *fileOperatorChannel) OpenFiles(folderPath string) <-chan UserStream {
	stream := make(chan UserStream)

	p, err := filepath.Abs(folderPath)
	if err != nil {
		panic(err)
	}

	paths := make([]string, 0)

	if err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		panic(fmt.Errorf("FileOperator: walk inside directory %s: %w", folderPath, err))
	}

	go func() {
		for _, p := range paths {
			t, err := os.Open(p)
			if err != nil {
				log.Printf("Error: file %s: err:%v\n", p, err)
				return
			}
			stream <- UserStream{file: t}
		}
		close(stream)
	}()

	return stream
}

func (f *fileOperatorChannel) Read(in <-chan UserStream) <-chan UserStream {
	out := make(chan UserStream, 5) // after 5 items the channel blocks

	go func() {
		for stream := range in {
			r := csv.NewReader(stream.file)
			r.Comment = '#'
			r.Comma = ','

			d, err := r.ReadAll()
			if err != nil {
				return
			}

			users := make(Users, 0)

			for i, l := range d {
				if i == 0 {
					continue
				}
				id, _ := strconv.Atoi(l[0])
				td, _ := time.Parse("2006/01/02", l[6])

				users = append(users, User{
					ID:        id,
					FirstName: l[1],
					LastName:  l[2],
					Email:     l[3],
					Gender:    l[4],
					Country:   l[5],
					Birthday:  td,
				})
			}

			out <- UserStream{
				file:  stream.file,
				users: users,
			}
		}
		close(out)
	}()

	return out
}

func (f *fileOperatorChannel) Write(done chan struct{}, in <-chan UserStream) {
	w := csv.NewWriter(f.writeFile)
	d := make([][]string, 0)

	go func() {
		defer close(done)
		for stream := range in {
			for _, user := range stream.users {
				d = append(d, user.ToArray())
			}
			if err := w.WriteAll(d); err != nil {
				log.Printf("error writing file with data %v: %v\n", d, err)
			}
		}
		done <- struct{}{}
	}()
}
