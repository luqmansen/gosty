package services

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/luqmansen/gosty/mock/pkg/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/mock/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/mock/pkg/apiserver/service"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/pkg/errors"
	"github.com/r3labs/sse/v2"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

type taskMatcher struct {
	x *models.Task
}

func (e *taskMatcher) Matches(x interface{}) bool {
	task2, ok := x.(*models.Task)
	if !ok {
		return false
	}

	e.x.TaskSubmitted = time.Time{}
	task2.TaskSubmitted = time.Time{}

	return reflect.DeepEqual(task2, e.x)
}

func (e *taskMatcher) String() string {
	return fmt.Sprintf("is equal to %v", e.x)
}

func EqTask(task *models.Task) gomock.Matcher {
	return &taskMatcher{x: task}
}

func Test_schedulerServices_CreateSplitTask(t *testing.T) {
	type fields struct {
		taskRepo  *mock_mongo.MockTaskRepository
		videoRepo repositories.VideoRepository
		messenger *mock_rabbitmq.MockMessenger
		sse       *sse.Server
		scheduler *mock_scheduler.MockScheduler
	}
	type args struct {
		video *models.Video
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		wantErr bool
	}{
		{
			name: "success create split task with size more than minimum size, expect error nil",
			prepare: func(f *fields) {
				_ = os.Setenv("FILE_MIN_SIZE_MB", "50")
				minSize := int64(50)
				video := &models.Video{Size: 10240 << 15}
				task := models.Task{
					OriginVideo: video,
					Kind:        models.TaskSplit,
					TaskSplit: &models.SplitTask{
						Video:       video,
						TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
						SizePerVid:  minSize,
						SizeLeft:    video.Size % minSize,
					},
					PrevTask:      models.TaskNew,
					Status:        models.TaskQueued,
					TaskSubmitted: time.Now(),
				}

				gomock.InOrder(
					f.taskRepo.EXPECT().Add(EqTask(&task)).Return(nil),
					f.messenger.EXPECT().Publish(EqTask(&task), MessageBrokerQueueTaskNew).Return(nil),
				)
			},
			args: args{
				video: &models.Video{Size: 10240 << 15},
			},
			wantErr: false,
		},
		{
			name: "success create split task with size less minimum size, expect error nil",
			prepare: func(f *fields) {
				video := &models.Video{Size: 1}
				task := &models.Task{
					OriginVideo:   video,
					TaskTranscode: &models.TranscodeTask{Video: video},
				}
				f.scheduler.EXPECT().CreateTranscodeTask(task).Return(nil)
			},
			args: args{
				video: &models.Video{Size: 1},
			},
			wantErr: false,
		},
		{
			name: "failed create split task with size less minimum size, expect error from CreateTranscodeTask",
			prepare: func(f *fields) {
				f.scheduler.EXPECT().CreateTranscodeTask(gomock.Any()).Return(errors.New("error happen"))
			},
			args: args{
				video: &models.Video{Size: 1},
			},
			wantErr: true,
		},
		{
			name: "failed to add task to task repository",
			prepare: func(f *fields) {
				f.taskRepo.EXPECT().Add(gomock.Any()).Return(errors.New("error happen"))
			},
			args: args{
				video: &models.Video{Size: 10240 << 15},
			},
			wantErr: true,
		},
		{
			name: "failed to publish task via messenger",
			prepare: func(f *fields) {
				gomock.InOrder(
					f.taskRepo.EXPECT().Add(gomock.Any()).Return(nil),
					f.messenger.EXPECT().Publish(gomock.Any(), MessageBrokerQueueTaskNew).Return(errors.New("error happen")),
				)
			},
			args: args{
				video: &models.Video{Size: 10240 << 15},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			defer controller.Finish()
			f := fields{
				taskRepo:  mock_mongo.NewMockTaskRepository(controller),
				messenger: mock_rabbitmq.NewMockMessenger(controller),
				scheduler: mock_scheduler.NewMockScheduler(controller),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}
			s := schedulerServices{
				taskRepo:  f.taskRepo,
				messenger: f.messenger,
			}

			if err := s.createSplitTask(tt.args.video, f.scheduler); (err != nil) != tt.wantErr {
				t.Errorf("CreateSplitTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_schedulerServices_createTranscodeTaskFromSplitTask(t *testing.T) {
	type fields struct {
		taskRepo  *mock_mongo.MockTaskRepository
		videoRepo repositories.VideoRepository
		messenger *mock_rabbitmq.MockMessenger
		sse       *sse.Server
		scheduler *mock_scheduler.MockScheduler
	}
	type args struct {
		task *models.Task
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		wantErr bool
	}{
		{
			name: "error happen when create task",
			prepare: func(f *fields) {
				f.scheduler.EXPECT().CreateTranscodeTask(gomock.Any()).Return(errors.New("error happen")).AnyTimes()
			},
			args: args{
				task: &models.Task{
					OriginVideo: &models.Video{},
					TaskSplit: &models.SplitTask{
						SplitedVideo: []*models.Video{{}},
					},
				}},
			wantErr: true,
		},
		{
			name: "success create task from split task",
			prepare: func(f *fields) {
				f.scheduler.EXPECT().CreateTranscodeTask(gomock.Any()).Return(nil).AnyTimes()
			},
			args: args{
				task: &models.Task{
					OriginVideo: &models.Video{},
					TaskSplit: &models.SplitTask{
						SplitedVideo: []*models.Video{{}},
					},
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		controller := gomock.NewController(t)
		defer controller.Finish()
		f := fields{
			taskRepo:  mock_mongo.NewMockTaskRepository(controller),
			scheduler: mock_scheduler.NewMockScheduler(controller),
		}
		if tt.prepare != nil {
			tt.prepare(&f)
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := schedulerServices{
				taskRepo:  f.taskRepo,
				videoRepo: f.videoRepo,
				messenger: f.messenger,
			}
			if err := s.createTranscodeTaskFromSplitTask(tt.args.task, f.scheduler); (err != nil) != tt.wantErr {
				t.Errorf("createTranscodeTaskFromSplitTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
