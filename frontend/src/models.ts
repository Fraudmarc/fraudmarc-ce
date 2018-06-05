export interface Job {
  action_type: string;
  answer: string;
  description: string;
  estimated_time: string;
  hit_name: string;
  job_id: string;
  start_time: string;
  status: string;
  steps: Step[];
}

export interface Step {
  image_url: string;
  step_name: string;
  title: string;
}
